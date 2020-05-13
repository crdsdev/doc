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
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	flag "github.com/spf13/pflag"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
)

var redisClient *redis.Client

// redis connection
var (
	envAddress   = "REDIS_HOST"
	envAnalytics = "ANALYTICS"

	address   string
	analytics bool = false
)

// SchemaPlusParent is a JSON schema plus the name of the parent field.
type SchemaPlusParent struct {
	Parent string
	Schema map[string]apiextensions.JSONSchemaProps
}

var docTemplate = template.Must(template.New("doc.html").Funcs(
	template.FuncMap{
		"plusParent": func(p string, s map[string]apiextensions.JSONSchemaProps) *SchemaPlusParent {
			return &SchemaPlusParent{
				Parent: p,
				Schema: s,
			}
		},
	},
).ParseFiles("template/doc.html", "template/analytics.html", "template/twitter.html"))

var orgTemplate = template.Must(template.ParseFiles("template/org.html", "template/analytics.html", "template/twitter.html"))
var newTemplate = template.Must(template.ParseFiles("template/new.html", "template/analytics.html", "template/twitter.html"))
var homeTemplate = template.Must(template.ParseFiles("template/home.html", "template/analytics.html", "template/twitter.html"))

type docData struct {
	Analytics   bool
	Repo        string
	Tag         string
	At          string
	Group       string
	Version     string
	Kind        string
	Description string
	Schema      apiextensions.JSONSchemaProps
}

type orgData struct {
	Analytics bool
	Repo      string
	Tag       string
	At        string
	CRDs      map[string]string
	Total     int
}

type homeData struct {
	Analytics bool
	Repos     []string
}

type newData struct {
	Analytics bool
}

func init() {
	address = os.Getenv(envAddress)

	// TODO(hasheddan): use a flag
	analyticsStr := os.Getenv(envAnalytics)
	if analyticsStr == "true" {
		analytics = true
	}
}

func main() {
	flag.Parse()

	redisClient = redis.NewClient(&redis.Options{
		Addr: address + ":6379",
	})
	start()
}

func start() {
	log.Println("Starting Doc server...")
	r := mux.NewRouter().StrictSlash(true)
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/", home)
	r.PathPrefix("/static/").Handler(staticHandler)
	r.HandleFunc("/github.com/{org}/{repo}@{tag}", org)
	r.HandleFunc("/github.com/{org}/{repo}", org)
	r.PathPrefix("/").HandlerFunc(doc)
	log.Fatal(http.ListenAndServe(":5000", r))
}

func home(w http.ResponseWriter, r *http.Request) {
	data := homeData{Analytics: analytics}
	if res, err := redisClient.SMembers("repos:popular").Result(); err != nil {
		log.Printf("failed to get popular repos : %v", err)
	} else {
		data.Repos = res
	}
	if err := homeTemplate.Execute(w, data); err != nil {
		log.Printf("homeTemplate.Execute(): %v", err)
		fmt.Fprint(w, "Unable to render home template.")
		return
	}
	log.Print("successfully rendered home page")
}

func org(w http.ResponseWriter, r *http.Request) {
	parameters := mux.Vars(r)
	org := parameters["org"]
	repo := parameters["repo"]
	tag := parameters["tag"]
	at := ""
	if tag != "" {
		at = "@"
	}
	res, err := redisClient.Get(strings.Join([]string{"github.com", org, repo}, "/") + at + tag).Result()
	if err != nil {
		log.Printf("failed to get CRDs for %s : %v", repo, err)
		if err := newTemplate.Execute(w, newData{Analytics: analytics}); err != nil {
			log.Printf("newTemplate.Execute(): %v", err)
			fmt.Fprint(w, "Unable to render new template.")
		}
		return
	}

	crds := &map[string]string{}
	bytes := []byte(res)
	if err := json.Unmarshal(bytes, crds); err != nil {
		log.Printf("failed to get CRDs for %s : %v", repo, err)
		http.ServeFile(w, r, "template/home.html")
		return
	}
	if err := orgTemplate.Execute(w, orgData{
		Analytics: analytics,
		Repo:      strings.Join([]string{org, repo}, "/"),
		Tag:       tag,
		At:        at,
		CRDs:      *crds,
		Total:     len(*crds),
	}); err != nil {
		log.Printf("orgTemplate.Execute(): %v", err)
		fmt.Fprint(w, "Unable to render org template.")
		return
	}
	log.Printf("successfully rendered org template")
}

func doc(w http.ResponseWriter, r *http.Request) {
	var schema *apiextensions.CustomResourceValidation
	crd := &apiextensions.CustomResourceDefinition{}
	log.Printf("Request Received: %s\n", r.URL.Path)
	org, repo, tag, err := parseGHURL(r.URL.Path)
	if err != nil {
		log.Printf("failed to parse Github path: %v", err)
		fmt.Fprint(w, "Invalid URL.")
		return
	}
	at := ""
	if tag != "" {
		at = "@"
	}
	res, err := redisClient.Get(strings.Trim(r.URL.Path, "/")).Result()
	if err != nil {
		log.Printf("failed to get CRDs for %s : %v", repo, err)
		if err := newTemplate.Execute(w, newData{Analytics: analytics}); err != nil {
			log.Printf("newTemplate.Execute(): %v", err)
			fmt.Fprint(w, "Unable to render new template.")
		}
		return
	}

	if err := json.Unmarshal([]byte(res), crd); err != nil {
		log.Printf("failed to convert to CRD: %v", err)
		fmt.Fprint(w, "Supplied file is not a valid CRD.")
		return
	}

	schema = crd.Spec.Validation
	if len(crd.Spec.Versions) > 1 {
		for _, version := range crd.Spec.Versions {
			if version.Storage == true {
				if version.Schema != nil {
					schema = version.Schema
				}
				break
			}
		}
	}

	if schema == nil || schema.OpenAPIV3Schema == nil {
		log.Print("CRD schema is nil.")
		fmt.Fprint(w, "Supplied CRD has no schema.")
		return
	}

	if err := docTemplate.Execute(w, docData{
		Analytics:   analytics,
		Repo:        strings.Join([]string{org, repo}, "/"),
		Tag:         tag,
		At:          at,
		Group:       crd.Spec.Group,
		Version:     crd.Spec.Version,
		Kind:        crd.Spec.Names.Kind,
		Description: string(schema.OpenAPIV3Schema.Description),
		Schema:      *schema.OpenAPIV3Schema,
	}); err != nil {
		log.Printf("docTemplate.Execute(): %v", err)
		fmt.Fprint(w, "Supplied CRD has no schema.")
		return
	}
	log.Printf("successfully rendered doc template")
}

// TODO(hasheddan): add testing and more reliable parse
func parseGHURL(uPath string) (org, repo, tag string, err error) {
	u, err := url.Parse(uPath)
	if err != nil {
		return "", "", "", err
	}
	elements := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(elements) < 4 {
		return "", "", "", errors.New("invalid path")
	}

	tagSplit := strings.Split(u.Path, "@")
	if len(tagSplit) > 1 {
		tag = tagSplit[1]
	}

	return elements[1], elements[2], tag, nil
}
