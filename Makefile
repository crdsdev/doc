# Set the shell to bash always
SHELL := /bin/bash

build-doc:
	docker build . -f deploy/doc.Dockerfile -t sijoma/doc:latest

build-gitter:
	docker build . -f deploy/gitter.Dockerfile -t sijoma/doc-gitter:latest

.PHONY: build-doc build-gitter
