#!/bin/bash

export GOPATH="${PWD}"
export GO111MODULE="off"

##go get -v -u -u ./...
go get -v -u github.com/alexflint/go-arg
go get -v -u github.com/anacrolix/dht
go get -v -u github.com/anacrolix/missinggo/httptoo
go get -v -u github.com/anacrolix/torrent
go get -v -u github.com/anacrolix/torrent/iplist
go get -v -u github.com/anacrolix/torrent/metainfo
go get -v -u github.com/anacrolix/utp
go get -v -u github.com/gin-gonic/gin
go get -v -u github.com/pion/webrtc/v2
go get -v -u go.etcd.io/bbolt
go get -v -u github.com/labstack/gommon
go get -v -u github.com/labstack/echo
go get -v -u github.com/dgrijalva/jwt-go
