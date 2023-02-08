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

package models

// RepoCRD is a CRD and data about its location in a repository.
type RepoCRD struct {
	Path     string
	Filename string
	Group    string
	Version  string
	Kind     string
	CRD      []byte
}

// GitterRepo is the repo for gitter to index.
type GitterRepo struct {
	Server string
	Org    string
	Repo   string
	Tag    string
}
