package main

import (
	"encoding/binary"
	"github.com/bradfitz/gomemcache/memcache"
)

type Memcached struct {
	client *memcache.Client
}

func NewMemcached(config *MemcachedConfig) *Memcached {
	client := memcache.New(config.addr...);
	
	return &Memcached{
		client: client,
	}
}

func (mem *Memcached) get(key string) (uint64, error) {
	item, err := mem.client.Get(key)

	if err != nil {
		return 0, err
	}

	return binary.LittleEndian.Uint64(item.Value), err
}

func (mem *Memcached) incrementByOne(key string) (uint64, error) {
	return mem.client.Increment(key, 1)
}

func (mem *Memcached) upsertObject(key string) (uint64, error) {
	
	_, err := mem.get(key)

	if err != nil {
		value := make([]byte, 8)
		binary.LittleEndian.PutUint64(value, 0) // initial value of 0
		item := &memcache.Item{
			Key: key,
			Value: value,
		}

		err := mem.client.Add(item)
		return 0, err

	} else {
		return mem.incrementByOne(key)
	}
}