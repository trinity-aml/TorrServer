package utils

import (
	"encoding/base32"
	"errors"
	"fmt"
	"context"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/anacrolix/torrent"
	"golang.org/x/time/rate"
)

var defTrackers = []string{
	"http://retracker.local",

	"http://bt4.t-ru.org/ann?magnet",
	"http://retracker.mgts.by:80/announce",
	"http://tracker.city9x.com:2710/announce",
	"http://tracker.electro-torrent.pl:80/announce",
	"http://tracker.internetwarriors.net:1337/announce",
	"http://tracker2.itzmx.com:6961/announce",
	"udp4://46.148.18.250:2710",
	"udp://opentor.org:2710",
	"udp://public.popcorn-tracker.org:6969/announce",
	"udp://tracker.opentrackr.org:1337/announce",

	"http://bt.svao-ix.ru/announce",

	"udp://explodie.org:6969/announce",

	//https://github.com/ngosang/trackerslist/blob/master/trackers_best_ip.txt 01.02.2020
	"udp://151.80.120.112:2710/announce",
	"udp://93.158.213.92:1337/announce",
	"udp://138.255.103.83:1337/announce",
	"udp://194.182.165.153:6969/announce",
	"udp://208.83.20.20:6969/announce",
	"udp://184.105.151.164:6969/announce",
	"udp://79.111.118.59:60889/announce",
	"http://93.157.234.32:60889/announce",
	"udp://51.81.46.170:6969/announce",
	"udp://51.68.199.47:6969/announce",
	"udp://5.206.60.196:6969/announce",
	"udp://185.181.60.67:80/announce",
	"udp://91.216.110.52:451/announce",
	"udp://89.234.156.205:451/announce",
	"udp://5.226.148.20:6969/announce",
	"udp://37.235.174.46:2710/announce",
	"udp://138.201.150.56:6969/announce",
}

var loadedTrackers []string

func GetDefTrackers() []string {
	loadNewTracker()
	if len(loadedTrackers) == 0 {
		return defTrackers
	}
	return loadedTrackers
}

func loadNewTracker() {
	if len(loadedTrackers) > 0 {
		return
	}
	resp, err := http.Get("https://newtrackon.com/api/stable")
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
			loadedTrackers = ret
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

func DnsResolve(host string, serverDNS string) (int, string)  {
	addrs, err := net.LookupHost(host)
	addr_dns := fmt.Sprintf("%s:53", serverDNS)
	a := 2
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
	b := fmt.Sprintf("Check dns: %s %s %s", host, addrs, err)
	return a,b
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
