#!/bin/bash

export CGO_ENABLED=0
export GOOS=linux
export GOARCH=amd64
go env -w GO111MODULE=on
go env -w GOPROXY=https://goproxy.cn,direct
go mod tidy
go mod download

cd ./gate
./build.sh
cd ..
cd ./hall
./build.sh
cd ..
cd ./fish
./build.sh

echo "build finish"

