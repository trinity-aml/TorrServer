package torr

import (
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"sync"
	"time"

	"server/settings"
	"server/torr/storage/memcache"
	"server/torr/storage/state"
	"server/utils"

	"log"

	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
)

type BTServer struct {
	config *torrent.ClientConfig
	client *torrent.Client

	storage *memcache.Storage

	torrents map[metainfo.Hash]*Torrent

	mu  sync.Mutex
	wmu sync.Mutex

	watching bool
}

func NewBTS() *BTServer {
	bts := new(BTServer)
	bts.torrents = make(map[metainfo.Hash]*Torrent)
	return bts
}

func (bt *BTServer) Connect() error {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	var err error
	bt.configure()
	bt.client, err = torrent.NewClient(bt.config)
	bt.torrents = make(map[metainfo.Hash]*Torrent)
	return err
}

func (bt *BTServer) Disconnect() {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	if bt.client != nil {
		bt.client.Close()
		bt.client = nil
		utils.FreeOSMemGC()
	}
}

func (bt *BTServer) Reconnect() error {
	bt.Disconnect()
	return bt.Connect()
}

func mediana(a float64, b float64) int64 {
	ret := int64(math.Round(a/b) * b)
	return ret
}

func (bt *BTServer) configure() {
	settings.Get().CacheSize = mediana(float64(settings.Get().CacheSize), float64(16*1024*1024))
	NewPreload := mediana(float64(settings.Get().PreloadBufferSize), float64(16*1024*1024))
	if NewPreload < 32*1024*1024 {
		NewPreload = 32 * 1024 * 1024
	}
	settings.Get().PreloadBufferSize = int64(NewPreload)
	bt.storage = memcache.NewStorage(settings.Get().CacheSize)

	blocklist, _ := utils.ReadBlockedIP()

	userAgent := "qBittorrent/4.3.2"
	peerID := "-qB4320-"
	cliVers := userAgent

	bt.config = torrent.NewDefaultClientConfig()

	bt.config.Debug = settings.Get().EnableDebug
	bt.config.DisableIPv6 = settings.Get().EnableIPv6 == false
	bt.config.DisableTCP = settings.Get().DisableTCP
	bt.config.DisableUTP = settings.Get().DisableUTP
	bt.config.NoDefaultPortForwarding = settings.Get().DisableUPNP
	bt.config.NoDHT = settings.Get().DisableDHT
	bt.config.NoUpload = settings.Get().DisableUpload
	bt.config.HeaderObfuscationPolicy = torrent.HeaderObfuscationPolicy{
		RequirePreferred: settings.Get().Encryption == 2, // Whether the value of Preferred is a strict requirement
		Preferred:        settings.Get().Encryption != 1, // Whether header obfuscation is preferred
	}
	bt.config.IPBlocklist = blocklist
	bt.config.DefaultStorage = bt.storage
	bt.config.Bep20 = peerID
	bt.config.PeerID = utils.PeerIDRandom(peerID)
	bt.config.HTTPUserAgent = userAgent
	bt.config.ExtendedHandshakeClientVersion = cliVers
	bt.config.EstablishedConnsPerTorrent = settings.Get().ConnectionsLimit
	bt.config.UpnpID = "YouROK/TorrServer_v1"
	timeout := settings.Get().TimeStrategy
	switch {
	case settings.Get().ChooseStrategy == 1:
		bt.config.DefaultRequestStrategy = torrent.RequestStrategyFastest()
	case settings.Get().ChooseStrategy == 2:
		bt.config.DefaultRequestStrategy = torrent.RequestStrategyFuzzing()
	case settings.Get().ChooseStrategy == 0:
		bt.config.DefaultRequestStrategy = torrent.RequestStrategyDuplicateRequestTimeout(timeout * time.Second)
	}
	if settings.Get().DhtConnectionLimit > 0 {
		bt.config.ConnTracker.SetMaxEntries(settings.Get().DhtConnectionLimit)
	} else if settings.Get().DhtConnectionLimit == 0 {
		if !settings.Get().DisableDHT {
			bt.config.TorrentPeersHighWater = 500
		}
	}
	if !settings.Get().DisableDHT {
		if settings.Get().ConnectionsLimit <= 20 {
			bt.config.TorrentPeersLowWater = settings.Get().ConnectionsLimit
		} else {
			bt.config.TorrentPeersLowWater = 50
		}
	}
	if int(float64(settings.Get().ConnectionsLimit)*0.5) <= 25 {
		bt.config.HalfOpenConnsPerTorrent = int(float64(settings.Get().ConnectionsLimit) * 0.5)
	} else {
		bt.config.HalfOpenConnsPerTorrent = 25
	}
	if settings.Get().DownloadRateLimit > 0 {
		bt.config.DownloadRateLimiter = utils.Limit(settings.Get().DownloadRateLimit * 1024)
	}
	if settings.Get().UploadRateLimit > 0 {
		bt.config.UploadRateLimiter = utils.Limit(settings.Get().UploadRateLimit * 1024)
	}
	if settings.Get().PeersListenPort > 0 {
		bt.config.ListenPort = settings.Get().PeersListenPort
		settings.Get().AutoListenPortN = 0
		settings.Get().AutoListenPort = false
	} else if settings.Get().PeersListenPort <= 0 && settings.Get().AutoListenPortN == 0 {
		if settings.Get().DisableUPNP == true {
			bt.config.NoDefaultPortForwarding = false
			settings.Get().DisableUPNP = false
		}
		for {
			m := 0
			a := 1024
			b := 32786
			rand.Seed(time.Now().UnixNano())
			n := a + rand.Intn(b-a+1)
			port := fmt.Sprintf(":%d", n)
			l, err := net.Listen("tcp", port)
			defer l.Close()
			if err == nil {
				bt.config.ListenPort = n
				log.Println("Open peers listen port:", n)
				m = 1
			} else {
				log.Println("Error:", err)
			}
			if m == 1 {
				if settings.Get().AutoListenPort == false {
					settings.Get().PeersListenPort = n
					settings.Get().AutoListenPortN = 0
				} else {
					settings.Get().AutoListenPortN = n
				}
				break
			}
		}
	}

	log.Println("Configure client:", settings.Get())
}

func (bt *BTServer) AddTorrent(magnet metainfo.Magnet, infobytes []byte, onAdd func(*Torrent)) (*Torrent, error) {
	torr, err := NewTorrent(magnet, infobytes, bt)
	if err != nil {
		return nil, err
	}

	if onAdd != nil {
		go func() {
			if torr.GotInfo() {
				onAdd(torr)
			}
		}()
	} else {
		go torr.GotInfo()
	}

	return torr, nil
}

func (bt *BTServer) List() []*Torrent {
	bt.mu.Lock()
	defer bt.mu.Unlock()
	list := make([]*Torrent, 0)
	for _, t := range bt.torrents {
		list = append(list, t)
	}
	return list
}

func (bt *BTServer) GetTorrent(hash metainfo.Hash) *Torrent {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	if t, ok := bt.torrents[hash]; ok {
		return t
	}

	return nil
}

func (bt *BTServer) RemoveTorrent(hash torrent.InfoHash) {
	if torr, ok := bt.torrents[hash]; ok {
		torr.Close()
	}
}

func (bt *BTServer) BTState() *BTState {
	bt.mu.Lock()
	defer bt.mu.Unlock()

	btState := new(BTState)
	btState.LocalPort = bt.client.LocalPort()
	btState.PeerID = fmt.Sprintf("%x", bt.client.PeerID())
	btState.BannedIPs = len(bt.client.BadPeerIPs())
	//	for _, dht := range bt.client.DhtServers() {
	//		btState.DHTs = append(btState.DHTs, dht)
	//	}
	for _, t := range bt.torrents {
		btState.Torrents = append(btState.Torrents, t)
	}
	return btState
}

func (bt *BTServer) CacheState(hash metainfo.Hash) *state.CacheState {
	st := bt.GetTorrent(hash)
	if st == nil {
		return nil
	}

	cacheState := bt.storage.GetStats(hash)
	return cacheState
}

func (bt *BTServer) WriteState(w io.Writer) {
	bt.client.WriteStatus(w)
}
