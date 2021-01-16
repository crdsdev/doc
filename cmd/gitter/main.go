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
	"os"
	"path"
	"regexp"
	"time"

	"github.com/crdsdev/doc/pkg/crd"
	"github.com/crdsdev/doc/pkg/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/jackc/pgx/v4"
	"gopkg.in/square/go-jose.v2/json"
)

const (
	crdArgCount = 7

	userEnv     = "PG_USER"
	passwordEnv = "PG_PASS"
	hostEnv     = "PG_HOST"
	portEnv     = "PG_PORT"
	dbEnv       = "PG_DB"
)

var (
	// TODO(hasheddan): temporarily hard-coded
	repos = []string{"crossplane/crossplane"}
)

func main() {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cloneDir := fmt.Sprintf("%s/%s", dir, "tmp-git")
	if err := os.Mkdir(cloneDir, os.FileMode(0755)); err != nil {
		panic(err)
	}
	defer os.RemoveAll(cloneDir)

	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", os.Getenv(userEnv), os.Getenv(passwordEnv), os.Getenv(hostEnv), os.Getenv(portEnv), os.Getenv(dbEnv))
	conn, err := pgx.Connect(context.Background(), dsn)
	if err != nil {
		panic(err)
	}

	if err := gitter(repos, conn); err != nil {
		panic(err)
	}
}

func gitter(repos []string, conn *pgx.Conn) error {
	log.Println("Starting gitter...")
	for _, repoURL := range repos {
		log.Printf("Indexing repo %s...\n", repoURL)
		dir, err := ioutil.TempDir(os.TempDir(), "doc-gitter")
		if err != nil {
			return err
		}
		defer os.RemoveAll(dir)
		repo, err := git.PlainClone(dir, false, &git.CloneOptions{
			URL: fmt.Sprintf("https://github.com/%s", repoURL),
		})
		if err != nil {
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
		if err := iter.ForEach(func(obj *plumbing.Reference) error {
			h, err := repo.ResolveRevision(plumbing.Revision(obj.Hash().String()))
			if err != nil || h == nil {
				log.Printf("Unable to resolve revision: %s (%v)", obj.Hash().String(), err)
				return nil
			}
			repoCrds, err := getCRDsFromTag(repoURL, dir, obj.Name().Short(), h, w)
			if err != nil {
				log.Printf("Unable to get CRDs: %s@%s (%v)", repoURL, obj.Name().Short(), err)
				return nil
			}
			if len(repoCrds.CRDs) > 0 {
				allArgs := make([]interface{}, 0, len(repoCrds.CRDs)*crdArgCount)
				for _, crd := range repoCrds.CRDs {
					allArgs = append(allArgs, crd.Group, crd.Version, crd.Kind, repoCrds.GithubURL, repoCrds.Tag, crd.Filename, crd.CRD)
				}
				if _, err := conn.Exec(context.Background(), buildInsert("INSERT INTO crds(\"group\", version, kind, repo, tag, filename, data) VALUES ", crdArgCount, len(repoCrds.CRDs)), allArgs...); err != nil {
					panic(err)
				}
			}
			return nil
		}); err != nil {
			log.Printf("Failed indexing %s, continuing to other repos: %v", repoURL, err)
			continue
		}
	}

	return nil
}

func getCRDsFromTag(repo string, dir string, tag string, hash *plumbing.Hash, w *git.Worktree) (*models.Repo, error) {
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
	repoData := &models.Repo{
		GithubURL:  "github.com" + "/" + repo,
		Tag:        tag,
		LastParsed: time.Now(),
		CRDs:       map[string]models.RepoCRD{},
	}
	files := splitYAML(g, dir)
	fileCount := 0
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
			repoData.CRDs[crd.PrettyGVK(crder.GVK)] = models.RepoCRD{
				Path:     crd.PrettyGVK(crder.GVK),
				Filename: path.Base(file),
				Group:    crder.GVK.Group,
				Version:  crder.GVK.Version,
				Kind:     crder.GVK.Kind,
				CRD:      cbytes,
			}
		}
		fileCount++
	}
	return repoData, nil
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
