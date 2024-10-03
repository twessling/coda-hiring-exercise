package main

import (
	"mrbarrel/router/router"
)

func main() {

	routerConfig := &router.Config{}

	router := router.New(routerConfig)

	router.ListenAndServe()
}
