#!/bin/bash

export GOPATH="${PWD}"
export GO111MODULE="off"

go get -v github.com/alexflint/go-arg@latest
go get -v github.com/anacrolix/dht@latest
go get -v github.com/anacrolix/missinggo/httptoo@latest
go get -v github.com/anacrolix/torrent@latest
go get -v github.com/anacrolix/torrent/iplist@latest
go get -v github.com/anacrolix/torrent/metainfo@latest
go get -v github.com/anacrolix/utp@latest
go get -u github.com/gin-gonic/gin@latest
go get -v github.com/pion/webrtc/v2@latest
go get -v go.etcd.io/bbolt@latest
go get -v github.com/labstack/gommon@latest
go get -v github.com/labstack/echo@latest
go get -v github.com/dgrijalva/jwt-go@latest
ln -s . src/github.com/pion/webrtc/v2
go get -v github.com/pion/webrtc/v2
