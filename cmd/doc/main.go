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
	"net/rpc"
	"net/url"
	"os"
	"strings"

	crdutil "github.com/crdsdev/doc/pkg/crd"
	"github.com/crdsdev/doc/pkg/models"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	flag "github.com/spf13/pflag"
	"github.com/unrolled/render"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"
)

var db *pgxpool.Pool

// redis connection
var (
	envAnalytics   = "ANALYTICS"
	envDevelopment = "IS_DEV"

	userEnv     = "PG_USER"
	passwordEnv = "PG_PASS"
	hostEnv     = "PG_HOST"
	portEnv     = "PG_PORT"
	dbEnv       = "PG_DB"

	cookieDarkMode = "halfmoon_preferredMode"

	address   string
	analytics bool = false

	gitterChan chan models.GitterRepo
)

// SchemaPlusParent is a JSON schema plus the name of the parent field.
type SchemaPlusParent struct {
	Parent string
	Schema map[string]apiextensions.JSONSchemaProps
}

var page = render.New(render.Options{
	Extensions:    []string{".html"},
	Directory:     "template",
	Layout:        "layout",
	IsDevelopment: os.Getenv(envDevelopment) == "true",
	Funcs: []template.FuncMap{
		{
			"plusParent": func(p string, s map[string]apiextensions.JSONSchemaProps) *SchemaPlusParent {
				return &SchemaPlusParent{
					Parent: p,
					Schema: s,
				}
			},
		},
	},
})

type pageData struct {
	Analytics     bool
	DisableNavBar bool
	IsDarkMode    bool
	Title         string
}

type baseData struct {
	Page pageData
}

type docData struct {
	Page        pageData
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
	Page   pageData
	Server string
	Repo   string
	Tag    string
	At     string
	Tags   []string
	CRDs   map[string]models.RepoCRD
	Total  int
}

type homeData struct {
	Page  pageData
	Repos []string
}

func worker(gitterChan <-chan models.GitterRepo) {
	for job := range gitterChan {
		client, err := rpc.DialHTTP("tcp", "127.0.0.1:1234")
		if err != nil {
			log.Fatal("dialing:", err)
		}
		reply := ""
		if err := client.Call("Gitter.Index", job, &reply); err != nil {
			log.Print("Could not index repo")
		}
	}
}

func tryIndex(repo models.GitterRepo, gitterChan chan models.GitterRepo) bool {
	select {
	case gitterChan <- repo:
		return true
	default:
		return false
	}
}

func init() {
	// TODO(hasheddan): use a flag
	analyticsStr := os.Getenv(envAnalytics)
	if analyticsStr == "true" {
		analytics = true
	}

	gitterChan = make(chan models.GitterRepo, 4)
}

func main() {
	flag.Parse()
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s", os.Getenv(userEnv), os.Getenv(passwordEnv), os.Getenv(hostEnv), os.Getenv(portEnv), os.Getenv(dbEnv))
	conn, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		panic(err)
	}
	db, err = pgxpool.ConnectConfig(context.Background(), conn)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 4; i++ {
		go worker(gitterChan)
	}

	start()
}

func getPageData(r *http.Request, title string, disableNavBar bool) pageData {
	var isDarkMode = false
	if cookie, err := r.Cookie(cookieDarkMode); err == nil && cookie.Value == "dark-mode" {
		isDarkMode = true
	}
	return pageData{
		Analytics:     analytics,
		IsDarkMode:    isDarkMode,
		DisableNavBar: disableNavBar,
		Title:         title,
	}
}

func start() {
	log.Println("Starting Doc server...")
	r := mux.NewRouter().StrictSlash(true)
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir("./static/")))
	r.HandleFunc("/", home)
	r.PathPrefix("/static/").Handler(staticHandler)
	r.HandleFunc("/{server}/{org}/{repo}@{tag}", org)
	r.HandleFunc("/{server}/{org}/{repo}", org)
	r.HandleFunc("/raw/{server}/{org}/{repo}@{tag}", raw)
	r.HandleFunc("/raw/{server}/{org}/{repo}", raw)
	r.PathPrefix("/").HandlerFunc(doc)
	log.Fatal(http.ListenAndServe(":5000", r))
}

func home(w http.ResponseWriter, r *http.Request) {
	data := homeData{Page: getPageData(r, "Doc", true)}
	if err := page.HTML(w, http.StatusOK, "home", data); err != nil {
		log.Printf("homeTemplate.Execute(): %v", err)
		fmt.Fprint(w, "Unable to render home template.")
		return
	}
	log.Print("successfully rendered home page")
}

func raw(w http.ResponseWriter, r *http.Request) {
	parameters := mux.Vars(r)
	server := parameters["server"]
	org := parameters["org"]
	repo := parameters["repo"]
	tag := parameters["tag"]

	fullRepo := fmt.Sprintf("%s/%s/%s", server, org, repo)
	var rows pgx.Rows
	var err error
	if tag == "" {
		rows, err = db.Query(context.Background(), "SELECT c.data::jsonb FROM tags t INNER JOIN crds c ON (c.tag_id = t.id) WHERE LOWER(t.repo)=LOWER($1) AND t.id = (SELECT id FROM tags WHERE LOWER(repo) = LOWER($1) ORDER BY time DESC LIMIT 1);", fullRepo)
	} else {
		rows, err = db.Query(context.Background(), "SELECT c.data::jsonb FROM tags t INNER JOIN crds c ON (c.tag_id = t.id) WHERE LOWER(t.repo)=LOWER($1) AND t.name=$2;", fullRepo, tag)
	}

	var res []byte
	var total []byte
	for err == nil && rows.Next() {
		if err := rows.Scan(&res); err != nil {
			break
		}
		crd := &apiextensions.CustomResourceDefinition{}
		if err := yaml.Unmarshal(res, crd); err != nil {
			break
		}
		crdv1 := &v1.CustomResourceDefinition{}
		if err := v1.Convert_apiextensions_CustomResourceDefinition_To_v1_CustomResourceDefinition(crd, crdv1, nil); err != nil {
			break
		}
		crdv1.SetGroupVersionKind(v1.SchemeGroupVersion.WithKind("CustomResourceDefinition"))
		y, err := yaml.Marshal(crdv1)
		if err != nil {
			break
		}
		total = append(total, y...)
		total = append(total, []byte("\n---\n")...)
	}

	if err != nil {
		fmt.Fprint(w, "Unable to render raw CRDs.")
		log.Printf("failed to get raw CRDs for %s : %v", repo, err)
	} else {
		w.Write([]byte(total))
		log.Printf("successfully rendered raw CRDs")
	}

	if analytics {
		u := uuid.New().String()
		// TODO(hasheddan): do not hardcode tid and dh
		metrics := url.Values{
			"v":   {"1"},
			"t":   {"pageview"},
			"tid": {"UA-116820283-2"},
			"cid": {u},
			"dh":  {"doc.crds.dev"},
			"dp":  {r.URL.Path},
			"uip": {r.RemoteAddr},
		}
		client := &http.Client{}

		req, _ := http.NewRequest("POST", "http://www.google-analytics.com/collect", strings.NewReader(metrics.Encode()))
		req.Header.Add("User-Agent", r.UserAgent())
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		if _, err := client.Do(req); err != nil {
			log.Printf("failed to report analytics: %s", err.Error())
		} else {
			log.Printf("successfully reported analytics")
		}
	}
}

func org(w http.ResponseWriter, r *http.Request) {
	parameters := mux.Vars(r)
	server := parameters["server"]
	org := parameters["org"]
	repo := parameters["repo"]
	tag := parameters["tag"]
	pageData := getPageData(r, fmt.Sprintf("%s/%s", org, repo), false)
	fullRepo := fmt.Sprintf("%s/%s/%s", server, org, repo)
	b := &pgx.Batch{}
	if tag == "" {
		b.Queue("SELECT t.name, c.group, c.version, c.kind FROM tags t INNER JOIN crds c ON (c.tag_id = t.id) WHERE LOWER(t.repo)=LOWER($1) AND t.id = (SELECT id FROM tags WHERE LOWER(repo) = LOWER($1) ORDER BY time DESC LIMIT 1);", fullRepo)
	} else {
		pageData.Title += fmt.Sprintf("@%s", tag)
		b.Queue("SELECT t.name, c.group, c.version, c.kind FROM tags t INNER JOIN crds c ON (c.tag_id = t.id) WHERE LOWER(t.repo)=LOWER($1) AND t.name=$2;", fullRepo, tag)
	}
	b.Queue("SELECT name FROM tags WHERE LOWER(repo)=LOWER($1) ORDER BY time DESC;", fullRepo)
	br := db.SendBatch(context.Background(), b)
	defer br.Close()
	c, err := br.Query()
	if err != nil {
		log.Printf("failed to get CRDs for %s : %v", repo, err)
		if err := page.HTML(w, http.StatusOK, "new", baseData{Page: pageData}); err != nil {
			log.Printf("newTemplate.Execute(): %v", err)
			fmt.Fprint(w, "Unable to render new template.")
		}
		return
	}
	repoCRDs := map[string]models.RepoCRD{}
	foundTag := tag
	for c.Next() {
		var t, g, v, k string
		if err := c.Scan(&t, &g, &v, &k); err != nil {
			log.Printf("newTemplate.Execute(): %v", err)
			fmt.Fprint(w, "Unable to render new template.")
		}
		foundTag = t
		repoCRDs[g+"/"+v+"/"+k] = models.RepoCRD{
			Group:   g,
			Version: v,
			Kind:    k,
		}
	}
	c, err = br.Query()
	if err != nil {
		log.Printf("failed to get tags for %s : %v", repo, err)
		if err := page.HTML(w, http.StatusOK, "new", baseData{Page: pageData}); err != nil {
			log.Printf("newTemplate.Execute(): %v", err)
			fmt.Fprint(w, "Unable to render new template.")
		}
		return
	}
	tags := []string{}
	tagExists := false
	for c.Next() {
		var t string
		if err := c.Scan(&t); err != nil {
			log.Printf("newTemplate.Execute(): %v", err)
			fmt.Fprint(w, "Unable to render new template.")
		}
		if !tagExists && t == tag {
			tagExists = true
		}
		tags = append(tags, t)
	}
	if len(tags) == 0 || (!tagExists && tag != "") {
		tryIndex(models.GitterRepo{
			Server: server,
			Org:    org,
			Repo:   repo,
			Tag:    tag,
		}, gitterChan)
		if err := page.HTML(w, http.StatusOK, "new", baseData{Page: pageData}); err != nil {
			log.Printf("newTemplate.Execute(): %v", err)
			fmt.Fprint(w, "Unable to render new template.")
		}
		return
	}
	if foundTag == "" {
		foundTag = tags[0]
	}
	if err := page.HTML(w, http.StatusOK, "org", orgData{
		Page:   pageData,
		Server: server,
		Repo:   strings.Join([]string{org, repo}, "/"),
		Tag:    foundTag,
		Tags:   tags,
		CRDs:   repoCRDs,
		Total:  len(repoCRDs),
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
	server, org, repo, group, kind, version, tag, err := parseGHURL(r.URL.Path)
	if err != nil {
		log.Printf("failed to parse Github path: %v", err)
		fmt.Fprint(w, "Invalid URL.")
		return
	}
	pageData := getPageData(r, fmt.Sprintf("%s.%s/%s", kind, group, version), false)
	fullRepo := fmt.Sprintf("%s/%s/%s", server, org, repo)
	var c pgx.Row
	if tag == "" {
		c = db.QueryRow(context.Background(), "SELECT t.name, c.data::jsonb FROM tags t INNER JOIN crds c ON (c.tag_id = t.id) WHERE LOWER(t.repo)=LOWER($1) AND t.id = (SELECT id FROM tags WHERE repo = $1 ORDER BY time DESC LIMIT 1) AND c.group=$2 AND c.version=$3 AND c.kind=$4;", fullRepo, group, version, kind)
	} else {
		c = db.QueryRow(context.Background(), "SELECT t.name, c.data::jsonb FROM tags t INNER JOIN crds c ON (c.tag_id = t.id) WHERE LOWER(t.repo)=LOWER($1) AND t.name=$2 AND c.group=$3 AND c.version=$4 AND c.kind=$5;", fullRepo, tag, group, version, kind)
	}
	foundTag := tag
	if err := c.Scan(&foundTag, crd); err != nil {
		log.Printf("failed to get CRDs for %s : %v", repo, err)
		if err := page.HTML(w, http.StatusOK, "doc", baseData{Page: pageData}); err != nil {
			log.Printf("newTemplate.Execute(): %v", err)
			fmt.Fprint(w, "Unable to render new template.")
		}
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

	gvk := crdutil.GetStoredGVK(crd)
	if gvk == nil {
		log.Print("CRD GVK is nil.")
		fmt.Fprint(w, "Supplied CRD has no GVK.")
		return
	}

	if err := page.HTML(w, http.StatusOK, "doc", docData{
		Page:        pageData,
		Repo:        strings.Join([]string{org, repo}, "/"),
		Tag:         foundTag,
		Group:       gvk.Group,
		Version:     gvk.Version,
		Kind:        gvk.Kind,
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
func parseGHURL(uPath string) (server, org, repo, group, version, kind, tag string, err error) {
	u, err := url.Parse(uPath)
	if err != nil {
		return "", "", "", "", "", "", "", err
	}
	elements := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(elements) < 6 {
		return "", "", "", "", "", "", "", errors.New("invalid path")
	}

	tagSplit := strings.Split(u.Path, "@")
	if len(tagSplit) > 1 {
		tag = tagSplit[1]
	}

	return elements[0], elements[1], elements[2], elements[3], elements[4], strings.Split(elements[5], "@")[0], tag, nil
}
