#!/bin/bash

export GOPATH="${PWD}"

##go get -v -u ./...
go get -v -u github.com/anacrolix/dht
go get -v -u github.com/anacrolix/missinggo
go get -v -u github.com/anacrolix/torrent
go get -v -u github.com/anacrolix/utp
