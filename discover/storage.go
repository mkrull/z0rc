package main

import (
	"encoding/json"
	"hash/fnv"
	"sync"
)

const shards = 32

// StorageBackend describes how to access and manage data in the data store
type StorageBackend interface {
	Put(string, []byte)
	Get(string) ([]byte, bool)
	Delete(string) ([]byte, bool)
	Has(string) bool
	Dump() ([]byte, error)
}

type inMemoryShard struct {
	Entries map[string][]byte
	sync.RWMutex
}

// InMemoryStore is a simple in memory storage backend
type InMemoryStore []*inMemoryShard

// NewInMemoryStore returns a reference to a simple in memory storage backend
func NewInMemoryStore() *InMemoryStore {
	s := make(InMemoryStore, shards)
	for i := 0; i < shards; i++ {
		s[i] = newInMemoryShard()
	}
	return &s
}

func newInMemoryShard() *inMemoryShard {
	return &inMemoryShard{
		Entries: make(map[string][]byte),
	}
}

func (s *InMemoryStore) getShard(key string) *inMemoryShard {
	h := fnv.New32()
	_, _ = h.Write([]byte(key))
	return (*s)[uint(h.Sum32())%uint(shards)]
}

// Put stores the value with the given key.
func (s *InMemoryStore) Put(key string, value []byte) {
	sh := s.getShard(key)
	sh.Lock()
	defer sh.Unlock()

	sh.Entries[key] = value
}

// Get retrieves the value with the given key.
func (s *InMemoryStore) Get(key string) ([]byte, bool) {
	sh := s.getShard(key)
	sh.Lock()
	defer sh.Unlock()

	value, ok := sh.Entries[key]

	return value, ok
}

// Has returns true if the given key exists.
func (s *InMemoryStore) Has(key string) bool {
	sh := s.getShard(key)
	sh.Lock()
	defer sh.Unlock()

	_, ok := sh.Entries[key]

	return ok
}

// Delete deletes the value with the given key.
func (s *InMemoryStore) Delete(key string) ([]byte, bool) {
	sh := s.getShard(key)
	sh.Lock()
	defer sh.Unlock()

	value, ok := sh.Entries[key]

	delete(sh.Entries, key)

	return value, ok
}

// Dump the full data store.
func (s *InMemoryStore) Dump() ([]byte, error) {
	entries := make(map[string]*json.RawMessage)

	for _, sh := range *s {
		sh.Lock()
		for k, v := range sh.Entries {
			val := json.RawMessage(v)
			entries[k] = &val
		}
		sh.Unlock()
	}

	buff, err := json.MarshalIndent(&entries, "", "  ")

	return buff, err
}
