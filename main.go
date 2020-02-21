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
	"context"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/google/go-github/v29/github"
	"github.com/gorilla/mux"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/util/rand"
)

var client *github.Client

var docTemplate = template.Must(template.New("doc.html").Funcs(
	template.FuncMap{
		"genRand": func() string {
			return rand.String(10)
		},
	},
).ParseFiles("doc.html"))

var orgTemplate = template.Must(template.ParseFiles("org.html"))

type docData struct {
	Repo        string
	Group       string
	Version     string
	Kind        string
	Description string
	Schema      apiextensions.JSONSchemaProps
}

type orgData struct {
	Repo  string
	CRDs  []github.CodeResult
	Total int
}

func start() {
	log.Println("Starting Doc server...")
	r := mux.NewRouter().StrictSlash(true)
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/", home)
	r.PathPrefix("/static/").Handler(staticHandler)
	r.HandleFunc("/github.com/{org}/{repo}", org)
	r.PathPrefix("/").HandlerFunc(doc)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func main() {
	client = github.NewClient(nil)
	start()
}

func home(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "home.html")
	log.Print("successfully rendered home page")
}

func org(w http.ResponseWriter, r *http.Request) {
	parameters := mux.Vars(r)
	org := parameters["org"]
	repo := parameters["repo"]
	query := fmt.Sprintf("q='kind: CustomResourceDefinition' in:file language:yaml repo:%s/%s", org, repo)
	code, _, err := client.Search.Code(context.TODO(), query, nil)
	if err != nil || code == nil {
		log.Printf("failed to get Github repo CRDs: %v", err)
		fmt.Fprintf(w, "Unable to find CRDs in %s/%s on Github.", org, repo)
		return
	}
	if err := orgTemplate.Execute(w, orgData{
		Repo:  strings.Join([]string{org, repo}, "/"),
		CRDs:  code.CodeResults,
		Total: code.GetTotal(),
	}); err != nil {
		log.Printf("orgTemplate.Execute(w, nil): %v", err)
		fmt.Fprint(w, "Unable to render org template.")
		return
	}
	log.Printf("successfully rendered org template")
}

func doc(w http.ResponseWriter, r *http.Request) {
	var schema *apiextensions.CustomResourceValidation
	log.Printf("Request Received: %s\n", r.URL.Path)
	org, repo, file, err := parseGHURL(r.URL.Path)
	if err != nil {
		log.Printf("failed to parse Github path: %v", err)
		fmt.Fprint(w, "Invalid URL.")
		return
	}
	contents, _, _, err := client.Repositories.GetContents(context.TODO(), org, repo, file, nil)
	if err != nil || contents == nil {
		log.Printf("failed to get Github contents: %v", err)
		fmt.Fprintf(w, "Unable to find file in %s/%s on Github.", org, repo)
		return
	}
	content, err := contents.GetContent()
	if err != nil {
		log.Printf("failed to get Github file contents: %v", err)
		fmt.Fprintf(w, "Unable to get file contents at path %s/%s/%s on Github.", org, repo, file)
		return
	}
	crder, err := NewCRDer([]byte(content), true)
	if err != nil || crder.crd == nil {
		log.Printf("failed to convert to CRD: %v", err)
		fmt.Fprint(w, "Supplied file is not a valid CRD.")
		return
	}

	schema = crder.crd.Spec.Validation
	if len(crder.crd.Spec.Versions) > 1 {
		for _, version := range crder.crd.Spec.Versions {
			if version.Storage == true {
				if version.Schema == nil {
					log.Printf("storage version has not schema")
					fmt.Fprint(w, "Specified storage version does not have a schema.")
					return
				}
				schema = version.Schema
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
		Repo:        strings.Join([]string{org, repo}, "/"),
		Group:       crder.crd.Spec.Group,
		Version:     crder.crd.Spec.Version,
		Kind:        crder.crd.Spec.Names.Kind,
		Description: string(schema.OpenAPIV3Schema.Description),
		Schema:      *schema.OpenAPIV3Schema,
	}); err != nil {
		log.Printf("docTemplate.Execute(w, nil): %v", err)
		fmt.Fprint(w, "Supplied CRD has no schema.")
		return
	}
	log.Printf("successfully rendered doc template")
}

// TODO(hasheddan): add testing and more reliable parse
func parseGHURL(uPath string) (org, repo, file string, err error) {
	u, err := url.Parse(uPath)
	if err != nil {
		return "", "", "", err
	}
	elements := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(elements) < 4 {
		return "", "", "", errors.New("invalid path")
	}

	return elements[1], elements[2], path.Join(elements[3:]...), nil
}
