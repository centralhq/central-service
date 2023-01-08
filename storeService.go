package main

type StoreService struct {
	store *Memcached
}

func NewStoreService(store *Memcached) *StoreService {
	return &StoreService{
		store: store,
	}
}

func (s *StoreService) upsert(key string) (uint64, error) {
	return s.store.upsertObject(key)
}