# -*- mode: makefile-gmake; coding: utf-8 -*-
#  # vi: set syntax=make :
# cSpell.language:en-GB
# # cSpell:disable
#

CLIENT_PROTO_DEFS = clientprotocol.proto
CLIENT_PROTO_TARGET = clientprotocol.pb.go
CLIENT_PROTO_GRPC_TARGET = clientprotocol_grpc.pb.go
CLIENT_PROTO_SWAGGER = clientprotocol.swagger.json

.PHONY: clean all

default: all

all: ${CLIENT_PROTO_TARGET} ${CLIENT_PROTO_SWAGGER} ${CLIENT_PROTO_GRPC_TARGET}

${CLIENT_PROTO_TARGET} ${CLIENT_PROTO_GRPC_TARGET}: ../protocols/${CLIENT_PROTO_DEFS}
	protoc -I../protocols \
		-I/usr/local/include -I. \
		-I${GOPATH}/src \
		--go_out=. \
		--go_opt=paths=source_relative \
		--go-grpc_out=. \
		--go-grpc_opt=paths=source_relative \
		../protocols/${CLIENT_PROTO_DEFS}

${CLIENT_PROTO_SWAGGER}: ../protocols/${CLIENT_PROTO_DEFS}
	protoc -I../protocols \
		-I/usr/local/include -I. \
		-I${GOPATH}/src \
  		--swagger_out=logtostderr=true:. \
		../protocols/${CLIENT_PROTO_DEFS}

clean: ## clean generated files
	-rm ${CLIENT_PROTO_TARGET}
	-rm ${CLIENT_PROTO_SWAGGER}
	-rm ${CLIENT_PROTO_GRPC_TARGET}
