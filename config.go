package main

type EngineConfig struct {
	PersistenceConfig *MemcachedConfig
}

func NewEngineConfig(persistenceConfig *MemcachedConfig) *EngineConfig {
	return &EngineConfig{persistenceConfig}
}