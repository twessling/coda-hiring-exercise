package main

import (
	"log"
	"mrbarrel/application/handler"
	"mrbarrel/application/registrator"
	"mrbarrel/lib/env"
)

func main() {
	// configure application
	addr := env.MustGetStringOrDefault("HTTP_ADDR", ":8080")
	handlerCfg := &handler.Config{
		Addr: addr,
	}

	routerConfig := &registrator.Config{
		RegistryAddr: env.MustGetString("REGISTRY_ADDR"),
		MyAddr:       addr,
	}

	handler := handler.New(handlerCfg)
	routerNotifier := registrator.New(routerConfig)

	// TODO: work with contexts to cancel things properly when this one dies
	go func() {
		err := handler.ListenAndServe()
		if err != nil {
			log.Fatalf("handler stopped: %v", err)
		}
	}()

	// TODO: can add a check to make sure handler is up before it registers itself to eliminate potential race conditions

	err := routerNotifier.Run()
	if err != nil {
		log.Fatalf("notify router stopped: %v", err)
	}
}
