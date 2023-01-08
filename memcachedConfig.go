package main

type MemcachedConfig struct {
	addr []string
}

func NewMemcachedConfig() *MemcachedConfig {
	return &MemcachedConfig{
		addr: []string{"10.0.0.1:11211"},
	}
}