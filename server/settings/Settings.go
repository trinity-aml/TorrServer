package settings

import (
	"encoding/json"
	"fmt"
	"math"
	"time"

	bolt "go.etcd.io/bbolt"
)

var (
	sets      *Settings
	StartTime time.Time
)

func init() {
	sets = new(Settings)
	sets.CacheSize = 192 * 1024 * 1024
	sets.EnableDebug = false
	sets.PreloadBufferSize = 32 * 1024 * 1024
	sets.ConnectionsLimit = 20
	sets.DhtConnectionLimit = 500
	sets.PeersListenPort = 0
	sets.AutoListenPort = true
	sets.AutoListenPortN = 0
	sets.RetrackersMode = 1
	sets.TorrentDisconnectTimeout = 30
	sets.ChooseStrategy = 0
	sets.TimeStrategy = 5
	sets.ChooseTrackers = 0
	StartTime = time.Now()
}

type Settings struct {
	CacheSize int64 // in byte, def 192 mb

	PreloadBufferSize int64 // in byte, buffer for preload

	RetrackersMode int //0 - don`t add, 1 - add retrackers, 2 - remove retrackers

	ChooseStrategy int //0 - default speed strategy (balanced), 1 - fast, 2 - fuzzing

	TimeStrategy time.Duration //5 - default timeout for default strategy

	ChooseTrackers int //0 - ngosang list of trackers, 1 - https://newtrackon.com

	//BT Config
	EnableIPv6               bool
	EnableDebug              bool
	DisableTCP               bool
	DisableUTP               bool
	DisableUPNP              bool
	DisableDHT               bool
	DisableUpload            bool
	ReadOnlyMode             bool
	Encryption               int // 0 - Enable, 1 - disable, 2 - force
	DownloadRateLimit        int // in kb, 0 - inf
	UploadRateLimit          int // in kb, 0 - inf
	ConnectionsLimit         int
	DhtConnectionLimit       int // 0 - inf
	PeersListenPort          int
	AutoListenPort           bool
	AutoListenPortN          int
	TorrentDisconnectTimeout int // in seconds
}

func Get() *Settings {
	return sets
}

func (s *Settings) String() string {
	buf, _ := json.MarshalIndent(sets, "", " ")
	return string(buf)
}

func mediana(a float64, b float64) int64 {
	ret := int64(math.Round(a/b) * b)
	return ret
}

func ReadSettings() error {
	err := openDB()
	if err != nil {
		return err
	}
	buf := make([]byte, 0)
	err = db.View(func(tx *bolt.Tx) error {
		sdb := tx.Bucket(dbSettingsName)
		if sdb == nil {
			return fmt.Errorf("error load settings")
		}

		buf = sdb.Get([]byte("json"))
		if buf == nil {
			return fmt.Errorf("error load settings")
		}
		return nil
	})
	err = json.Unmarshal(buf, sets)
	if err != nil {
		return err
	}
	if sets.ConnectionsLimit <= 0 {
		sets.ConnectionsLimit = 20
	}
	if sets.DhtConnectionLimit < 0 {
		sets.DhtConnectionLimit = 500
	}
	if sets.CacheSize < 0 {
		sets.CacheSize = 192 * 1024 * 1024
	}
	if sets.TorrentDisconnectTimeout < 1 {
		sets.TorrentDisconnectTimeout = 1
	}
	sets.CacheSize = mediana(float64(sets.CacheSize), float64(16*1024*1024))
	NewPreload := mediana(float64(sets.PreloadBufferSize), float64(16*1024*1024))
	if NewPreload < 32*1024*1024 {
		NewPreload = 32 * 1024 * 1024
	}
	sets.PreloadBufferSize = NewPreload
	sets.AutoListenPortN = 0
	return nil
}

func SaveSettings() error {
	err := openDB()
	if err != nil {
		return err
	}

	buf, err := json.Marshal(sets)
	if err != nil {
		return err
	}

	return db.Update(func(tx *bolt.Tx) error {
		setsDB, err := tx.CreateBucketIfNotExists(dbSettingsName)
		if err != nil {
			return err
		}
		return setsDB.Put([]byte("json"), []byte(buf))
	})
}

func SetRDB() {
	SaveSettings()
	fmt.Println("Enable Read-only DB mode")
	CloseDB()
	sets.ReadOnlyMode = true
}
