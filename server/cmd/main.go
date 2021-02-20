package main

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"server"
	"server/settings"
	"server/utils"
	"server/version"
)

type args struct {
	Port string `arg:"-p" help:"web server port"`
	Path string `arg:"-d" help:"database path"`
	Add  string `arg:"-a" help:"add torrent link and exit"`
	RDB  bool   `arg:"-r" help:"start in read-only DB mode"`
	Kill bool   `arg:"-k" help:"dont kill program on signal"`
}

func (args) Version() string {
	return "TorrServer " + version.Version
}

var params args

func main() {
	arg.MustParse(&params)

	if params.Path == "" {
		params.Path, _ = os.Getwd()
	}

	if params.Port == "" {
		params.Port = "8091"
	}

	if params.Add != "" {
		add()
	}

	hosts := [6]string{"1.1.1.1", "1.0.0.1", "208.67.222.222", "208.67.220.220", "8.8.8.8", "8.8.4.4"}
	ret := 0
	for _, ip := range hosts {
		ret = utils.DnsResolve("www.themoviedb.org",ip)
		switch {
			case ret == 2:
				fmt.Println("DNS resolver OK\n")
			case ret == 1:
				fmt.Println("New DNS resolver OK\n")
			case ret == 0:
				fmt.Println("New DNS resolver failed\n")
		}
		if ret == 2 || ret == 1 {
			break
		}
	}
	
	Preconfig(params.Kill)

	server.Start(params.Path, params.Port)
	if (params.RDB) {
	    settings.SetRDB()
	} else {
	    settings.SaveSettings()
	}
	fmt.Println(server.WaitServer())
	time.Sleep(time.Second * 3)
	os.Exit(0)
}

func add() {
	err := addRemote()
	if err != nil {
		fmt.Println("Error add torrent:", err)
		os.Exit(-1)
	}

	fmt.Println("Added ok")
	os.Exit(0)
}

func addRemote() error {
	url := "http://localhost:" + params.Port + "/torrent/add"
	fmt.Println("Add torrent link:", params.Add, "\n", url)

	json := `{"Link":"` + params.Add + `"}`
	resp, err := http.Post(url, "text/html; charset=utf-8", bytes.NewBufferString(json))
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	return nil
}
