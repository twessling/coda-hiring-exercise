package main

import (
	"context"
	"log"
	"mrbarrel/lib/env"
	"mrbarrel/lib/shutdown"
	"mrbarrel/router/handler"
	"mrbarrel/router/pool"
	"sync"
	"time"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	// configuration phase
	poolConfig := &pool.PoolConfig{
		MaxAgeNoNotif: env.MustGetDurationOrDefault("MAX_CLIENT_NO_NOTIF", time.Second*2),
	}
	poolHandlerConfig := &handler.RegistryHandlerConfig{
		ListenAddr: env.MustGetStringOrDefault("REGISTRY_ADDR", ":8081"),
	}
	routerConfig := &handler.RouterConfig{
		Addr: env.MustGetStringOrDefault("HTTP_ADDR", ":8081"),
	}

	// wiring phase
	clientPool, clientRegistrar := pool.NewPool(poolConfig)
	poolHandler := handler.NewRegistryHandler(poolHandlerConfig, clientRegistrar)
	router := handler.NewRouter(routerConfig, clientPool)

	// run phase
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		shutdown.ListenStopSignal(ctx, cancelFunc)
	}()

	go func() {
		defer wg.Done()
		clientPool.Run(ctx)
	}()

	go func() {
		defer wg.Done()
		err := poolHandler.ListenForClients(ctx)
		if err != nil {
			log.Printf("ERROR in clientPool: %v", err)
		}
		cancelFunc()
	}()

	go func() {
		defer wg.Done()
		err := router.ListenAndServe(ctx)
		if err != nil {
			log.Printf("ERROR in handler: %v", err)
		}
		cancelFunc()

	}()

	log.Print("Router service up and running.")
	wg.Wait()
	log.Print("Router service shutdown complete, exiting. May I rise again.")
}
