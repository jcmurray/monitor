# -*- mode: makefile-gmake; coding: utf-8 -*-
#  # vi: set syntax=make :
# cSpell.language:en-GB
# # cSpell:disable
#
.PHONY: all clientapi clean dep

default: all

all: clean clientapi dep monitor

dep: ## Get the dependencies
	@go get -v -d ./...

clientapi: ## generate the protocol stubs
	${MAKE} -C clientapi

monitor: clientapi dep
	go build ./...
	go build
	go vet ./...

clean: ## clean all generated files
	-rm monitor
	${MAKE} -C clientapi clean
