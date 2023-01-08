package main

type StoreResource interface {
	upsert(key string) (uint64, error)
}