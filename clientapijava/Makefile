# -*- mode: makefile-gmake; coding: utf-8 -*-
#  # vi: set syntax=make :
# cSpell.language:en-GB
# # cSpell:disable
#



CLIENT_PROTO_DEFS = clientprotocol.proto
CLIENT_PROTO_TARGET = ClientProtocol.java
CLIENT_PROTO_SWAGGER = clientprotocol.swagger.json

.PHONY: clean all

default: all

all: ${CLIENT_PROTO_TARGET} ${CLIENT_PROTO_SWAGGER}

${CLIENT_PROTO_TARGET}: ../protocols/${CLIENT_PROTO_DEFS}
	protoc -I../protocols \
		-I/usr/local/include -I. \
		--java_out=. \
		${CLIENT_PROTO_DEFS}

${CLIENT_PROTO_SWAGGER}: ../protocols/${CLIENT_PROTO_DEFS}
	protoc -I../protocols \
		-I/usr/local/include -I. \
  		--swagger_out=logtostderr=true:. \
		${CLIENT_PROTO_DEFS}

clean: ## clean generated files
	-rm -rf com
	-rm ${CLIENT_PROTO_SWAGGER}
