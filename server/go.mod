module server

go 1.16

require (
	github.com/alexflint/go-arg v1.2.0
	github.com/anacrolix/dht/v2 v2.5.1-0.20200317023935-129f05e9b752
	github.com/anacrolix/missinggo/v2 v2.4.1-0.20200227072623-f02f6484f997
	github.com/anacrolix/torrent v1.15.0
	github.com/labstack/echo/v4 v4.2.0
	github.com/labstack/gommon v0.3.0
	go.etcd.io/bbolt v1.3.4
	golang.org/x/time v0.0.0-20201208040808-7e3f01d25324
)

replace github.com/anacrolix/dht v0.0.0-20180412060941-24cbf25b72a4 => github.com/anacrolix/dht v1.0.1
