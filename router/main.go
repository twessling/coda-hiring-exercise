package main

import (
	"context"
	"log"
	"mrbarrel/lib/env"
	"mrbarrel/lib/shutdown"
	"mrbarrel/router/pool"
	"mrbarrel/router/router"
	"sync"
	"time"
)

func main() {
	ctx, cancelFunc := context.WithCancel(context.Background())

	poolConfig := &pool.Config{
		ListenAddr:    env.MustGetStringOrDefault("ROUTER_ADDR", ":8081"),
		MaxAgeNoNotif: env.MustGetDurationOrDefault("MAX_CLIENT_NO_NOTIF", time.Second*2),
	}
	routerConfig := &router.Config{
		Addr: env.MustGetStringOrDefault("HTTP_ADDR", ":8081"),
	}

	clientPool := pool.New(poolConfig)
	router := router.New(routerConfig, clientPool)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		shutdown.ListenStopSignal(ctx, cancelFunc)
	}()

	go func() {
		defer wg.Done()
		err := clientPool.ListenForClients(ctx)
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
