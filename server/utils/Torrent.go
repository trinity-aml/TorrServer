package utils

import (
	"context"
	"encoding/base32"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"server/settings"

	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"
)

var defTrackers = []string{
	"http://retracker.local/announce",
	"http://bt4.t-ru.org/ann?magnet",
	"http://retracker.mgts.by:80/announce",
	"http://tracker.city9x.com:2710/announce",
	"http://tracker.electro-torrent.pl:80/announce",
	"http://tracker.internetwarriors.net:1337/announce",
	"http://tracker2.itzmx.com:6961/announce",
	"udp://opentor.org:2710/announce",
	"udp://public.popcorn-tracker.org:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",
	"http://tracker.filetracker.pl:8089/announce",
	"http://tracker2.wasabii.com.tw:6969/announce",
	"http://tracker.grepler.com:6969/announce",
	"http://tracker.tiny-vps.com:6969/announce",
	"http://tracker.dler.org:6969/announce",
	"udp://ipv6.leechers-paradise.org:6969",
	"http://bt.svao-ix.ru/announce",
	"udp://explodie.org:6969/announce",
}

var loadedTrackers []string

func GetDefTrackers() []string {
	var hosts []string
	if settings.Get().ChooseTrackers == 0 {
		hosts = []string{"https://raw.githubusercontent.com/ngosang/trackerslist/master/trackers_best_ip.txt"}
		fmt.Println("Load ngosang announcers\n")
	} else if settings.Get().ChooseTrackers == 1 {
		hosts = []string{"https://newtrackon.com/api/stable"}
		fmt.Println("Load newtrackon.com announcers\n")
	}
	for _, ip := range hosts {
		loadNewTracker(ip)
	}
	if len(loadedTrackers) == 0 {
		fmt.Println("Load external announcers failed\n")
		fmt.Println("Loaded built-in announcers list\n")
		return defTrackers
	} else {
		fmt.Println("Load external announcers successed\n")
		return loadedTrackers
	}
}

func loadNewTracker(host string) {
	resp, err := http.Get(host)
	if err == nil {
		buf, err := ioutil.ReadAll(resp.Body)
		if err == nil {
			arr := strings.Split(string(buf), "\n")
			var ret []string
			for _, s := range arr {
				s = strings.TrimSpace(s)
				if len(s) > 0 {
					ret = append(ret, s)
				}
			}
			if len(loadedTrackers) == 0 {
				loadedTrackers = []string{"http://retracker.local/announce"}
				loadedTrackers = append(loadedTrackers, ret...)
			} else {
				loadedTrackers = append(loadedTrackers, ret...)
			}
		}
	}
}

func PeerIDRandom(peer string) string {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		panic(err)
	}
	return peer + base32.StdEncoding.EncodeToString(randomBytes)[:20-len(peer)]
}

func GotInfo(t *torrent.Torrent, timeout int) error {
	gi := t.GotInfo()
	select {
	case <-gi:
		return nil
	case <-time.Tick(time.Second * time.Duration(timeout)):
		return errors.New("timeout load torrent info")
	}
}

func DnsResolve(host string, serverDNS string) int {
	addrs, err := net.LookupHost(host)
	addr_dns := fmt.Sprintf("%s:53", serverDNS)
	a := 0
	if len(addrs) == 0 {
		fmt.Println("Check dns", addrs, err)
		fn := func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{}
			return d.DialContext(ctx, "udp", addr_dns)
		}
		net.DefaultResolver = &net.Resolver{
			Dial: fn,
		}
		addrs, err = net.LookupHost(host)
		fmt.Println("Check new dns", addrs, err)
		if err == nil || len(addrs) > 0 {
			a = 1
		} else {
			a = 0
		}
	} else {
		a = 2
	}
	return a
}

func Limit(i int) *rate.Limiter {
	l := rate.NewLimiter(rate.Inf, 0)
	if i > 0 {
		b := i
		if b < 16*1024 {
			b = 16 * 1024
		}
		l = rate.NewLimiter(rate.Limit(i), b)
	}
	return l
}
