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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/crdsdev/doc/pkg/crd"
	"github.com/go-redis/redis/v7"
	flag "github.com/spf13/pflag"
	"gopkg.in/square/go-jose.v2/json"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

var redisClient *redis.Client

var (
	envAddress = "REDIS_HOST"
	envRepos   = "REPOS"

	address string
	repos   []string
)

func init() {
	address = os.Getenv(envAddress)

	repos = strings.Split(os.Getenv(envRepos), ",")
}

func main() {
	flag.Parse()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	cloneDir := fmt.Sprintf("%s/%s", dir, "tmp-git")
	if err := os.Mkdir(cloneDir, os.FileMode(0755)); err != nil {
		panic(err)
	}
	defer os.RemoveAll(cloneDir)

	redisClient = redis.NewClient(&redis.Options{
		Addr: address + ":6379",
	})

	if err := gitter(repos, redisClient); err != nil {
		panic(err)
	}
}

func gitter(repos []string, r *redis.Client) error {
	log.Println("Starting gitter...")

	for _, repoURL := range repos {
		log.Printf("Indexing repo %s...\n", repoURL)
		dir, err := ioutil.TempDir("tmp", "doc-gitter")
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

		iter, err := repo.TagObjects()
		if err != nil {
			return err
		}

		w, err := repo.Worktree()
		if err != nil {
			return err
		}

		// Get master
		repoCrds, crds, err := getCRDsFromMaster(repoURL, dir, w)
		if err != nil {
			log.Printf("Unable to get CRDs: %s@%s", repoURL, "master")
			return err
		}
		if err := r.MSet(crds); err.Err() != nil {
			log.Printf("Unable to mass set CRD contents for %s@%s", repoURL, "master")
			return err.Err()
		}
		bytes, err := json.Marshal(repoCrds)
		if err != nil {
			return err
		}
		if err := r.Set("github.com"+"/"+repoURL, bytes, 0); err.Err() != nil {
			log.Printf("Unable to set CRD list for %s@%s", repoURL, "master")
			return err.Err()
		}

		// Get tags
		if err := iter.ForEach(func(obj *object.Tag) error {
			log.Printf(obj.Name)
			repoCrds, crds, err := getCRDsFromTag(repoURL, dir, obj, w)
			if err != nil {
				log.Printf("Unable to get CRDs: %s@%s", repoURL, obj.Name)
				return err
			}
			if err := r.MSet(crds); err.Err() != nil {
				log.Printf("Unable to mass set CRD contents for %s@%s", repoURL, obj.Name)
				return err.Err()
			}
			bytes, err := json.Marshal(repoCrds)
			if err != nil {
				return err
			}
			if err := r.Set("github.com"+"/"+repoURL+"@"+obj.Name, bytes, 0); err.Err() != nil {
				log.Printf("Unable to set CRD list for %s@%s", repoURL, obj.Name)
				return err.Err()
			}
			return nil
		}); err != nil {
			log.Printf("Failed indexing %s, continuing to other repos: %v", repoURL, err)
			continue
		}
	}

	return nil
}

func getCRDsFromTag(repo string, dir string, obj *object.Tag, w *git.Worktree) (map[string]string, map[string]interface{}, error) {
	commit, err := obj.Commit()
	if err != nil {
		return nil, nil, err
	}
	err = w.Checkout(&git.CheckoutOptions{
		Hash: commit.Hash,
	})
	if err != nil {
		return nil, nil, err
	}
	reg := regexp.MustCompile("kind: CustomResourceDefinition")
	regPath := regexp.MustCompile("^.*\\.yaml")
	g, _ := w.Grep(&git.GrepOptions{
		Patterns:  []*regexp.Regexp{reg},
		PathSpecs: []*regexp.Regexp{regPath},
	})
	repoCrds := map[string]string{}
	crds := map[string]interface{}{}
	for _, res := range g {
		b, err := ioutil.ReadFile(dir + "/" + res.FileName)
		if err != nil {
			log.Printf("failed to read CRD file: %s", res.FileName)
			continue
		}
		repoCrds[res.FileName] = path.Base(res.FileName)

		crder, err := crd.NewCRDer([]byte(b))
		if err != nil || crder.CRD == nil {
			log.Printf("failed to convert to CRD: %v", err)
			continue
		}

		bytes, err := json.Marshal(crder.CRD)
		if err != nil {
			log.Printf("failed to marshal CRD: %s/%s, %v", res.FileName, obj.Name, err)
			continue
		}
		crds["github.com"+"/"+repo+"/"+res.FileName+"@"+obj.Name] = bytes
	}
	return repoCrds, crds, nil
}

func getCRDsFromMaster(repo string, dir string, w *git.Worktree) (map[string]string, map[string]interface{}, error) {
	reg := regexp.MustCompile("kind: CustomResourceDefinition")
	regPath := regexp.MustCompile("^.*\\.yaml")
	g, _ := w.Grep(&git.GrepOptions{
		Patterns:  []*regexp.Regexp{reg},
		PathSpecs: []*regexp.Regexp{regPath},
	})
	repoCrds := map[string]string{}
	crds := map[string]interface{}{}
	for _, res := range g {
		b, err := ioutil.ReadFile(dir + "/" + res.FileName)
		if err != nil {
			log.Printf("failed to read CRD file: %s", res.FileName)
			continue
		}
		repoCrds[res.FileName] = path.Base(res.FileName)

		crder, err := crd.NewCRDer([]byte(b))
		if err != nil || crder.CRD == nil {
			log.Printf("failed to convert to CRD: %v", err)
			continue
		}

		bytes, err := json.Marshal(crder.CRD)
		if err != nil {
			log.Printf("failed to marshal CRD: %s/%s, %v", res.FileName, "master", err)
			continue
		}
		crds["github.com"+"/"+repo+"/"+res.FileName] = bytes
	}
	return repoCrds, crds, nil
}
