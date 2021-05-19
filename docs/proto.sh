#!/bin/env sh
protoc --go_out=../internal --go_opt=paths=source_relative --go-grpc_out=../internal --go-grpc_opt=paths=source_relative proto/*.proto
protoc --doc_out=resource/custom_markdown.tpl,index.md:./ proto/*.proto