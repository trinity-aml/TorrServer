module server

go 1.16

require (
	github.com/alexflint/go-arg v1.3.0
	github.com/anacrolix/dht/v2 v2.8.0
	github.com/anacrolix/missinggo/v2 v2.5.0
	github.com/anacrolix/torrent v1.25.0
	github.com/labstack/echo/v4 v4.2.0
	github.com/labstack/gommon v0.3.0
	go.etcd.io/bbolt v1.3.5
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
)

replace github.com/anacrolix/dht v0.0.0-20180412060941-24cbf25b72a4 => github.com/anacrolix/dht/v2 v2.5.0
