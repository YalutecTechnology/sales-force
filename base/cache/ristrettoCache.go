package cache

import (
	"encoding/binary"
	"time"

	"github.com/labstack/gommon/log"
	"yalochat.com/salesforce-integration/base/constants"

	"github.com/dgraph-io/ristretto"
)

type cacheRistretto struct {
	cacheLocal *ristretto.Cache
}

type ICache interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}, ttl time.Duration) bool
	Delete(key string)
	Wait()
	Clear()
}

func New() *cacheRistretto {

	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: constants.NumCounters,
		MaxCost:     constants.MaxCost,
		BufferItems: constants.BufferItems,
	})
	if err != nil {
		log.Fatalf("[cache could not be generated in ristretto][err:%s]", err)
	}

	return &cacheRistretto{
		cacheLocal: cache,
	}
}

func (r *cacheRistretto) Get(key string) (interface{}, bool) {
	return r.cacheLocal.Get(key)
}

func (r *cacheRistretto) Set(key string, value interface{}, ttl time.Duration) bool {
	return r.cacheLocal.SetWithTTL(key, value, int64(binary.Size(value)), ttl)
}

func (r *cacheRistretto) Delete(key string) {
	r.cacheLocal.Del(key)
}

func (r *cacheRistretto) Wait() {
	r.cacheLocal.Wait()
}

func (r *cacheRistretto) Clear() {
	r.cacheLocal.Clear()
}
