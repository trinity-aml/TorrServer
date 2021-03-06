package memcache

import (
	"log"
	"sort"
	"sync"

	"server/settings"
	"server/torr/reader"
	"server/torr/storage/state"
	"server/utils"

	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/storage"
)

type Cache struct {
	storage.TorrentImpl

	s *Storage

	capacity int64
	filled   int64
	hash     metainfo.Hash

	pieceLength int64
	pieceCount  int
	piecesBuff  int

	muPiece  sync.Mutex
	muRemove sync.Mutex
	muReader sync.Mutex
	isRemove bool

	pieces     map[int]*Piece
	bufferPull *BufferPool

	prcLoaded int

	readers map[*reader.Reader]struct{}
}

func NewCache(capacity int64, storage *Storage) *Cache {
	ret := &Cache{
		capacity: capacity,
		filled:   0,
		pieces:   make(map[int]*Piece),
		s:        storage,
		readers:  make(map[*reader.Reader]struct{}),
	}

	return ret
}

func (c *Cache) Init(info *metainfo.Info, hash metainfo.Hash) {
	log.Println("Create cache for:", info.Name)
	switch {
	case settings.Get().CacheSize > info.PieceLength*12:
		c.capacity = info.PieceLength * 12
	case settings.Get().CacheSize == 0:
		c.capacity = info.PieceLength * 12
	}
	c.pieceLength = info.PieceLength
	c.pieceCount = info.NumPieces()
	c.piecesBuff = int(c.capacity / c.pieceLength)
	c.hash = hash
	c.bufferPull = NewBufferPool(c.pieceLength, c.capacity)

	for i := 0; i < c.pieceCount; i++ {
		c.pieces[i] = &Piece{
			Id:     i,
			Length: info.Piece(i).Length(),
			Hash:   info.Piece(i).Hash().HexString(),
			cache:  c,
		}
	}
}

func (c *Cache) Piece(m metainfo.Piece) storage.PieceImpl {
	c.muPiece.Lock()
	defer c.muPiece.Unlock()
	if val, ok := c.pieces[m.Index()]; ok {
		return val
	}
	return nil
}

func (c *Cache) Close() error {
	c.isRemove = false
	log.Println("Close cache for:", c.hash)
	if _, ok := c.s.caches[c.hash]; ok {
		delete(c.s.caches, c.hash)
	}
	c.pieces = nil
	c.bufferPull = nil
	c.readers = nil
	utils.FreeOSMemGC()
	return nil
}

func (c *Cache) GetState() state.CacheState {
	cState := state.CacheState{}
	cState.Capacity = c.capacity
	cState.PiecesLength = c.pieceLength
	cState.PiecesCount = c.pieceCount
	cState.Hash = c.hash.HexString()

	stats := make(map[int]state.ItemState, 0)
	c.muPiece.Lock()
	var fill int64 = 0
	for _, value := range c.pieces {
		stat := value.Stat()
		if stat.BufferSize > 0 {
			fill += stat.BufferSize
			stats[stat.Id] = stat
		}
	}
	c.filled = fill
	c.muPiece.Unlock()
	cState.Filled = c.filled
	cState.Pieces = stats
	return cState
}

func (c *Cache) cleanPieces() {
	if c.isRemove {
		return
	}
	c.muRemove.Lock()
	if c.isRemove {
		c.muRemove.Unlock()
		return
	}
	c.isRemove = true
	defer func() { c.isRemove = false }()
	c.muRemove.Unlock()

	remPieces := c.getRemPieces()
	if len(remPieces) > 0 && (c.filled > c.capacity || c.bufferPull.Len() <= 1) {
		remCount := int((c.filled - c.capacity) / c.pieceLength)
		if remCount < 1 {
			remCount = 1
		}
		if remCount > len(remPieces) {
			remCount = len(remPieces)
		}

		remPieces = remPieces[:remCount]

		for _, p := range remPieces {
			c.removePiece(p)
		}
	}
}

func (c *Cache) getRemPieces() []*Piece {
	pieces := make([]*Piece, 0)
	fill := int64(0)
	loading := 0
	used := c.bufferPull.Used()
	for u := range used {
		v := c.pieces[u]
		if v.Size > 0 {
			if v.Id > 0 {
				pieces = append(pieces, v)
			}
			fill += v.Size
			if !v.complete {
				loading++
			}
		}
	}
	c.filled = fill
	sort.Slice(pieces, func(i, j int) bool {
		return pieces[i].accessed < pieces[j].accessed
	})

	c.prcLoaded = prc(c.piecesBuff-loading, c.piecesBuff)
	return pieces
}

func (c *Cache) removePiece(piece *Piece) {
	c.muPiece.Lock()
	defer c.muPiece.Unlock()
	piece.Release()

	if c.prcLoaded >= 75 {
		utils.FreeOSMemGC()
	} else {
		utils.FreeOSMem()
	}
}

func prc(val, of int) int {
	return int(float64(val) * 100.0 / float64(of))
}

func (c *Cache) AddReader(r *reader.Reader) {
	c.muReader.Lock()
	defer c.muReader.Unlock()
	c.readers[r] = struct{}{}
}

func (c *Cache) RemReader(r *reader.Reader) {
	c.muReader.Lock()
	defer c.muReader.Unlock()
	delete(c.readers, r)
}

func (c *Cache) ReadersLen() int {
	if c == nil || c.readers == nil {
		return 0
	}
	return len(c.readers)
}

func (c *Cache) AdjustRA(readahead int64) {
	c.muReader.Lock()
	defer c.muReader.Unlock()
	if settings.Get().CacheSize == 0 || settings.Get().CacheSize > readahead*3 {
		c.capacity = readahead * 3
	}
	for r, _ := range c.readers {
		r.SetReadahead(readahead)
		r.SetResponsive()
	}
}
