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
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"path"
	"regexp"
	"time"

	"github.com/crdsdev/doc/pkg/crd"
	"github.com/crdsdev/doc/pkg/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
	"gopkg.in/square/go-jose.v2/json"
)

const (
	crdArgCount = 6

	userEnv     = "PG_USER"
	passwordEnv = "PG_PASS"
	hostEnv     = "PG_HOST"
	portEnv     = "PG_PORT"
	dbEnv       = "PG_DB"
)

func main() {
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", os.Getenv(userEnv), os.Getenv(passwordEnv), os.Getenv(hostEnv), os.Getenv(portEnv), os.Getenv(dbEnv))
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
}

type tag struct {
	timestamp time.Time
	hash      plumbing.Hash
	name      string
}

// Index indexes a git repo at the specified url.
func (g *Gitter) Index(gRepo models.GitterRepo, reply *string) error {
	log.Printf("Indexing repo %s/%s...\n", gRepo.Org, gRepo.Repo)

	dir, err := ioutil.TempDir(os.TempDir(), "doc-gitter")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	cloneOpts := &git.CloneOptions{
		URL:               fmt.Sprintf("https://github.com/%s/%s", gRepo.Org, gRepo.Repo),
		Depth:             1,
		Progress:          os.Stdout,
		RecurseSubmodules: git.NoRecurseSubmodules,
	}
	if gRepo.Tag != "" {
		cloneOpts.ReferenceName = plumbing.NewTagReferenceName(gRepo.Tag)
		cloneOpts.SingleBranch = true
	}
	repo, err := git.PlainClone(dir, false, cloneOpts)
	if err != nil {
		return err
	}
	iter, err := repo.TagObjects()
	if err != nil {
		return err
	}
	w, err := repo.Worktree()
	if err != nil {
		return err
	}
	// Get CRDs for each tag
	tags := []tag{}
	if err := iter.ForEach(func(obj *object.Tag) error {
		if gRepo.Tag == "" {
			tags = append(tags, tag{
				timestamp: obj.Tagger.When,
				hash:      obj.Target,
				name:      obj.Name,
			})
			return nil
		}
		if obj.Name == gRepo.Tag {
			tags = append(tags, tag{
				timestamp: obj.Tagger.When,
				hash:      obj.Target,
				name:      obj.Name,
			})
			iter.Close()
		}
		return nil
	}); err != nil {
		log.Println(err)
	}
	for _, t := range tags {
		r := g.conn.QueryRow(context.Background(), "SELECT id FROM tags WHERE name=$1 AND repo=$2", t.name, "github.com/"+gRepo.Org+"/"+gRepo.Repo)
		var tagID int
		if err := r.Scan(&tagID); err != nil {
			if !errors.Is(err, pgx.ErrNoRows) {
				return err
			}
			r := g.conn.QueryRow(context.Background(), "INSERT INTO tags(name, repo, time) VALUES ($1, $2, $3) RETURNING id", t.name, "github.com/"+gRepo.Org+"/"+gRepo.Repo, t.timestamp)
			if err := r.Scan(&tagID); err != nil {
				return err
			}
		}
		h, err := repo.ResolveRevision(plumbing.Revision(t.hash.String()))
		if err != nil || h == nil {
			log.Printf("Unable to resolve revision: %s (%v)", t.hash.String(), err)
			continue
		}
		repoCRDs, err := getCRDsFromTag(gRepo.Org+"/"+gRepo.Repo, dir, t.name, h, w)
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

func getCRDsFromTag(repo string, dir string, tag string, hash *plumbing.Hash, w *git.Worktree) (map[string]models.RepoCRD, error) {
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
	regPath := regexp.MustCompile("^.*\\.yaml")
	g, _ := w.Grep(&git.GrepOptions{
		Patterns:  []*regexp.Regexp{reg},
		PathSpecs: []*regexp.Regexp{regPath},
	})
	repoCRDs := map[string]models.RepoCRD{}
	files := splitYAML(g, dir)
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

func splitYAML(greps []git.GrepResult, dir string) map[string][][]byte {
	allCRDs := map[string][][]byte{}
	for _, res := range greps {
		b, err := ioutil.ReadFile(dir + "/" + res.FileName)
		if err != nil {
			log.Printf("failed to read CRD file: %s", res.FileName)
			continue
		}
		// TODO(hasheddan): generalize this replacement
		b = bytes.ReplaceAll(b, []byte("rw-rw----"), []byte("660"))
		allCRDs[res.FileName] = bytes.Split(b, []byte("---"))
	}
	return allCRDs
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
