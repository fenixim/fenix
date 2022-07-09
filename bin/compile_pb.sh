#!/bin/sh

# Check if protoc is installed
if ! command -v protoc &> /dev/null
then
    echo "protoc not installed!"
    echo "protoc is required to compile protobufs. Please follow instructions on https://grpc.io/docs/protoc-installation/ to install."
    exit
fi

# Check if protoc-gen-go is installed
if ! command -v protoc-gen-go &> /dev/null
then
    echo "Trying to install protoc-gen-go"
    go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
    go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
    echo "Add export PATH='\$PATH:\$(go env GOPATH)/bin' to your profile."
    export PATH="$PATH:$(go env GOPATH)/bin"
fi
cd protobuf
# Make the protobufs
protoc --go_out=../src/pb --go_opt=paths=source_relative \
    --go-grpc_out=../src/pb --go-grpc_opt=paths=source_relative \
    *.proto