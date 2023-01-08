package main

import (
	"go.uber.org/dig"
)

func BuildContainer() *dig.Container {
	container := dig.New()

	container.Provide(NewMemcachedConfig)
	container.Provide(NewEngineConfig)
	container.Provide(NewMemcached)
	container.Provide(NewStoreService, dig.As(new(StoreResource)))
	container.Provide(NewOperationManager)
	container.Provide(NewEngineServer)
	container.Provide(NewHub)
	
	
	return container
} 

func main() {
	container := BuildContainer()

	err := container.Invoke(func(server *EngineServer, hub *Hub) {
		
		go hub.run()
		server.Run()
	})

	if err != nil {
		panic(err)
	}
}