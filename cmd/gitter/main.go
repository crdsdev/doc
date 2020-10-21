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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/crdsdev/doc/pkg/crd"
	"github.com/crdsdev/doc/pkg/models"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-redis/redis/v7"
	flag "github.com/spf13/pflag"
	"gopkg.in/square/go-jose.v2/json"
)

const popularReposSet = "repos:popular"

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

	if _, err := r.SAdd(popularReposSet, repos).Result(); err != nil {
		log.Printf("Failed to add repos to set: %v", err)
	}

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

		iter, err := repo.Tags()
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
		if err := iter.ForEach(func(obj *plumbing.Reference) error {
			h, err := repo.ResolveRevision(plumbing.Revision(obj.Hash().String()))
			if err != nil || h == nil {
				log.Printf("Unable to resolve revision: %s (%v)", obj.Hash().String(), err)
			}
			repoCrds, crds, err := getCRDsFromTag(repoURL, dir, obj.Name().Short(), h, w)
			if err != nil {
				log.Printf("Unable to get CRDs: %s@%s (%v)", repoURL, obj.Name().Short(), err)
				return nil
			}
			if len(crds) > 0 {
				if err := r.MSet(crds); err.Err() != nil {
					log.Printf("Unable to mass set CRD contents for %s@%s", repoURL, obj.Name().Short())
					return err.Err()
				}
			}
			bytes, err := json.Marshal(repoCrds)
			if err != nil {
				return err
			}
			if err := r.Set("github.com"+"/"+repoURL+"@"+obj.Name().Short(), bytes, 0); err.Err() != nil {
				log.Printf("Unable to set CRD list for %s@%s", repoURL, obj.Name().Short())
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

func getCRDsFromTag(repo string, dir string, tag string, hash *plumbing.Hash, w *git.Worktree) (*models.Repo, map[string]interface{}, error) {
	log.Printf("Getting CRDs for tag (%s) at commit (%s)", tag, hash.String())

	err := w.Checkout(&git.CheckoutOptions{
		Hash:  *hash,
		Force: true,
	})
	if err != nil {
		return nil, nil, err
	}
	if err := w.Reset(&git.ResetOptions{
		Mode: git.HardReset,
	}); err != nil {
		return nil, nil, err
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
	crds := map[string]interface{}{}
	files := splitYAML(g, dir)
	for file, yamls := range files {
		for _, y := range yamls {
			crder, err := crd.NewCRDer(y, crd.StripLabels(), crd.StripAnnotations(), crd.StripConversion())
			if err != nil || crder.CRD == nil {
				log.Printf("failed to convert to CRD: %v", err)
				continue
			}

			bytes, err := json.Marshal(crder.CRD)
			if err != nil {
				log.Printf("failed to marshal CRD in: %s/%s, %v", file, tag, err)
				continue
			}
			repoData.CRDs[crd.PrettyGVK(crder.GVK)] = models.RepoCRD{
				Path:     crd.PrettyGVK(crder.GVK),
				Filename: path.Base(file),
				Group:    crder.CRD.Spec.Group,
				Version:  crder.CRD.Spec.Version,
				Kind:     crder.CRD.Spec.Names.Kind,
			}
			crds["github.com"+"/"+repo+"/"+crd.PrettyGVK(crder.GVK)+"@"+tag] = bytes
		}
	}
	return repoData, crds, nil
}

func getCRDsFromMaster(repo string, dir string, w *git.Worktree) (*models.Repo, map[string]interface{}, error) {
	reg := regexp.MustCompile("kind: CustomResourceDefinition")
	regPath := regexp.MustCompile("^.*\\.yaml")
	g, _ := w.Grep(&git.GrepOptions{
		Patterns:  []*regexp.Regexp{reg},
		PathSpecs: []*regexp.Regexp{regPath},
	})
	repoData := &models.Repo{
		GithubURL:  "github.com" + "/" + repo,
		Tag:        "master",
		LastParsed: time.Now(),
		CRDs:       map[string]models.RepoCRD{},
	}
	crds := map[string]interface{}{}
	files := splitYAML(g, dir)
	for file, yamls := range files {
		for _, y := range yamls {
			crder, err := crd.NewCRDer(y, crd.StripLabels(), crd.StripAnnotations(), crd.StripConversion())
			if err != nil || crder.CRD == nil {
				log.Printf("failed to convert to CRD: %v", err)
				continue
			}

			bytes, err := json.Marshal(crder.CRD)
			if err != nil {
				log.Printf("failed to marshal CRD in: %s/%s, %v", file, "master", err)
				continue
			}

			repoData.CRDs[crd.PrettyGVK(crder.GVK)] = models.RepoCRD{
				Path:     crd.PrettyGVK(crder.GVK),
				Filename: path.Base(file),
				Group:    crder.CRD.Spec.Group,
				Version:  crder.CRD.Spec.Version,
				Kind:     crder.CRD.Spec.Names.Kind,
			}
			crds["github.com"+"/"+repo+"/"+crd.PrettyGVK(crder.GVK)] = bytes
		}
	}
	return repoData, crds, nil
}

func splitYAML(greps []git.GrepResult, dir string) map[string][][]byte {
	allCRDs := map[string][][]byte{}
	for _, res := range greps {
		b, err := ioutil.ReadFile(dir + "/" + res.FileName)
		if err != nil {
			log.Printf("failed to read CRD file: %s", res.FileName)
			continue
		}
		allCRDs[res.FileName] = bytes.Split(b, []byte("---"))
	}
	return allCRDs
}
