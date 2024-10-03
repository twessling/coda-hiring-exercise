package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"mrbarrel/application/handler"
	"mrbarrel/application/registrator"
	"mrbarrel/lib/env"
	"mrbarrel/lib/shutdown"
	"sync"
	"time"
)

func main() {
	// configure application phase
	host := env.MustGetStringOrDefault("HTTP_HOST", "")
	port := env.MustGetIntOrDefault("HTTP_PORT", 8080)
	idBytes := make([]byte, 8)
	if _, err := rand.Read(idBytes); err != nil {
		log.Fatalf("while creating random id: %v", err)
	}

	handlerCfg := &handler.Config{
		Addr: fmt.Sprintf("%s:%d", host, port),
		Id:   base64.StdEncoding.EncodeToString(idBytes),
	}

	routerConfig := &registrator.Config{
		RegistryAddr:  env.MustGetString("REGISTRY_ADDR"),
		MyAddr:        fmt.Sprintf("%s:%d", env.MustGetString("HOSTNAME"), port), // here we need the docker host name
		NotifInterval: env.MustGetDurationOrDefault("REGISTRY_INTERVAL", time.Second),
	}

	handler := handler.New(handlerCfg)
	routerNotifier := registrator.New(routerConfig)

	// run application phase

	ctx, cancelFunc := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		shutdown.ListenStopSignal(ctx, cancelFunc)
	}()

	go func() {
		defer wg.Done()
		err := handler.ListenAndServe(ctx)
		if err != nil {
			log.Printf("ERROR in handler: %v", err)
		}
		cancelFunc()
	}()

	// TODO: can add a check to make sure handler is up before it registers itself to eliminate potential race conditions

	go func() {
		defer wg.Done()
		err := routerNotifier.Run(ctx)
		if err != nil {
			log.Printf("ERROR in router notifier: %v", err)
		}
		cancelFunc()
	}()

	log.Printf("API service %s up and running.", handlerCfg.Id)
	wg.Wait()
	log.Print("API Service shutdown complete, exiting. May I rise again.")
}
