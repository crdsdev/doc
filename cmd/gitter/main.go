/*
Copyright 2020 The CRDS Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/crdsdev/doc/pkg/crd"
	"github.com/crdsdev/doc/pkg/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	gogithttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2/json"
	"gopkg.in/yaml.v3"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"regexp"
	"strings"
	"time"
)

const (
	crdArgCount = 6

	userEnv     = "PG_USER"
	passwordEnv = "PG_PASS"
	hostEnv     = "PG_HOST"
	portEnv     = "PG_PORT"
	dbEnv       = "PG_DB"
)

type Config struct {
	Repos []RepoConfig `yaml:"repos"`
}

func (c *Config) getAuthenticationUser(gRepo models.GitterRepo) *RepoUser {
	for _, element := range c.Repos {
		if element.Server == gRepo.Server && element.Name == gRepo.Org+"/"+gRepo.Repo {
			return &element.User
		}
	}
	return &RepoUser{Name: "", PwdFromEnv: ""}
}

type RepoConfig struct {
	Name   string   `yaml:"name"`
	Server string   `yaml:"server"`
	User   RepoUser `yaml:"user"`
}

type RepoUser struct {
	Name       string `yaml:"name"`
	PwdFromEnv string `yaml:"pwdFromEnv"`
}

func readConf(filename string) (*Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	c := &Config{}
	err = yaml.Unmarshal(buf, c)
	if err != nil {
		return nil, fmt.Errorf("in file %q: %w", filename, err)
	}

	return c, err
}

func main() {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", os.Getenv(userEnv), os.Getenv(passwordEnv), os.Getenv(hostEnv), os.Getenv(portEnv), os.Getenv(dbEnv))
	conf, err := readConf("/usr/repos.yaml")
	if err != nil {
		fmt.Println(err)
	}
	conn, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}
	pool, err := pgxpool.ConnectConfig(context.Background(), conn)
	if err != nil {
		panic(err)
	}
	gitter := &Gitter{
		conn: pool,
		conf: conf,
	}
	rpc.Register(gitter)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	log.Println("Starting gitter...")
	http.Serve(l, nil)
}

// Gitter indexes git repos.
type Gitter struct {
	conn *pgxpool.Pool
	conf *Config
}

type tag struct {
	timestamp time.Time
	hash      plumbing.Hash
	name      string
}

// Index indexes a git repo at the specified url.
func (g *Gitter) Index(gRepo models.GitterRepo, reply *string) error {
	dir, err := ioutil.TempDir(os.TempDir(), "doc-gitter")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	authUser := g.conf.getAuthenticationUser(gRepo)

	var fullRepo string

	fullRepo = fmt.Sprintf("%s/%s/%s", strings.ToLower(gRepo.Server), strings.ToLower(gRepo.Org), strings.ToLower(gRepo.Repo))

	log.Printf("Indexing repo %s...\n", fullRepo)

	cloneOpts := &git.CloneOptions{
		URL:               fmt.Sprintf("https://%s", fullRepo),
		Depth:             1,
		Progress:          os.Stdout,
		RecurseSubmodules: git.NoRecurseSubmodules,
		Auth: &gogithttp.BasicAuth{
			Username: authUser.Name,
			Password: os.Getenv(authUser.PwdFromEnv),
		},
	}
	if gRepo.Tag != "" {
		cloneOpts.ReferenceName = plumbing.NewTagReferenceName(gRepo.Tag)
		cloneOpts.SingleBranch = true
	}
	repo, err := git.PlainClone(dir, false, cloneOpts)
	if err != nil {
		fmt.Println(err)
		return err
	}
	iter, err := repo.Tags()
	if err != nil {
		return err
	}
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	// Get CRDs for each tag
	tags := []tag{}
	if err := iter.ForEach(func(obj *plumbing.Reference) error {
		if gRepo.Tag == "" {
			tags = append(tags, tag{
				hash: obj.Hash(),
				name: obj.Name().Short(),
			})
			return nil
		}
		if obj.Name().Short() == gRepo.Tag {
			tags = append(tags, tag{
				hash: obj.Hash(),
				name: obj.Name().Short(),
			})
			iter.Close()
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	for _, t := range tags {
		h, err := repo.ResolveRevision(plumbing.Revision(t.hash.String()))
		if err != nil || h == nil {
			log.Printf("Unable to resolve revision: %s (%v)", t.hash.String(), err)
			continue
		}
		c, err := repo.CommitObject(*h)
		if err != nil || c == nil {
			log.Printf("Unable to resolve revision: %s (%v)", t.hash.String(), err)
			continue
		}
		r := g.conn.QueryRow(context.Background(), "SELECT id FROM tags WHERE name=$1 AND repo=$2", t.name, fullRepo)
		var tagID int
		if err := r.Scan(&tagID); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			r := g.conn.QueryRow(context.Background(), "INSERT INTO tags(name, repo, time) VALUES ($1, $2, $3) RETURNING id", t.name, fullRepo, c.Committer.When)
			if err := r.Scan(&tagID); err != nil {
				return err
			}
		}
		repoCRDs, err := getCRDsFromTag(dir, t.name, h, w)
		if err != nil {
			log.Printf("Unable to get CRDs: %s@%s (%v)", repo, t.name, err)
			continue
		}
		if len(repoCRDs) > 0 {
			allArgs := make([]interface{}, 0, len(repoCRDs)*crdArgCount)
			for _, crd := range repoCRDs {
				allArgs = append(allArgs, crd.Group, crd.Version, crd.Kind, tagID, crd.Filename, crd.CRD)
			}
			if _, err := g.conn.Exec(context.Background(), buildInsert("INSERT INTO crds(\"group\", version, kind, tag_id, filename, data) VALUES ", crdArgCount, len(repoCRDs))+"ON CONFLICT DO NOTHING", allArgs...); err != nil {
				return err
			}
		}
	}

	log.Printf("Finished indexing %s/%s\n", gRepo.Org, gRepo.Repo)

	return nil
}

func getCRDsFromTag(dir string, tag string, hash *plumbing.Hash, w *git.Worktree) (map[string]models.RepoCRD, error) {
	err := w.Checkout(&git.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
	if err != nil {
		return nil, err
	}
	if err := w.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	}); err != nil {
		return nil, err
	}
	reg := regexp.MustCompile("kind: CustomResourceDefinition")
	regPath := regexp.MustCompile(`^.*\.yaml`)
	g, _ := w.Grep(&git.GrepOptions{
		Patterns:  []*regexp.Regexp{reg},
		PathSpecs: []*regexp.Regexp{regPath},
	})
	repoCRDs := map[string]models.RepoCRD{}
	files := getYAMLs(g, dir)
	for file, yamls := range files {
		for _, y := range yamls {
			crder, err := crd.NewCRDer(y, crd.StripLabels(), crd.StripAnnotations(), crd.StripConversion())
			if err != nil || crder.CRD == nil {
				continue
			}
			cbytes, err := json.Marshal(crder.CRD)
			if err != nil {
				continue
			}
			repoCRDs[crd.PrettyGVK(crder.GVK)] = models.RepoCRD{
				Path:     crd.PrettyGVK(crder.GVK),
				Filename: path.Base(file),
				Group:    crder.GVK.Group,
				Version:  crder.GVK.Version,
				Kind:     crder.GVK.Kind,
				CRD:      cbytes,
			}
		}
	}
	return repoCRDs, nil
}

func getYAMLs(greps []git.GrepResult, dir string) map[string][][]byte {
	allCRDs := map[string][][]byte{}
	for _, res := range greps {
		b, err := ioutil.ReadFile(dir + "/" + res.FileName)
		if err != nil {
			log.Printf("failed to read CRD file: %s", res.FileName)
			continue
		}

		yamls, err := splitYAML(b, res.FileName)
		if err != nil {
			log.Printf("failed to split/parse CRD file: %s", res.FileName)
			continue
		}

		allCRDs[res.FileName] = yamls
	}
	return allCRDs
}

func splitYAML(file []byte, filename string) ([][]byte, error) {
	var yamls [][]byte
	var err error = nil
	defer func() {
		if err := recover(); err != nil {
			yamls = make([][]byte, 0)
			err = fmt.Errorf("panic while processing yaml file: %w", err)
		}
	}()

	decoder := yaml.NewDecoder(bytes.NewReader(file))
	for {
		var node map[string]interface{}
		err := decoder.Decode(&node)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("failed to decode part of CRD file: %s\n%s", filename, err)
			continue
		}

		doc, err := yaml.Marshal(node)
		if err != nil {
			log.Printf("failed to encode part of CRD file: %s\n%s", filename, err)
			continue
		}
		yamls = append(yamls, doc)
	}
	return yamls, err
}

func buildInsert(query string, argsPerInsert, numInsert int) string {
	absArg := 1
	for i := 0; i < numInsert; i++ {
		query += "("
		for j := 0; j < argsPerInsert; j++ {
			query += "$" + fmt.Sprint(absArg)
			if j != argsPerInsert-1 {
				query += ","
			}
			absArg++
		}
		query += ")"
		if i != numInsert-1 {
			query += ","
		}
	}
	return query
}
