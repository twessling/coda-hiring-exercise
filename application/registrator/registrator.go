package registrator

import (
	"net/http"
	"strings"
	"time"
)

type Config struct {
	RegistryAddr  string
	MyAddr        string
	NotifInterval time.Duration
}

type Registrator struct {
	registryAddr string
	interval     time.Duration
	myAddr       string
}

func New(cfg *Config) *Registrator {
	return &Registrator{
		registryAddr: cfg.RegistryAddr,
		myAddr:       cfg.MyAddr,
		interval:     cfg.NotifInterval,
	}
}

func (r *Registrator) Run() error {
	for {
		body := strings.NewReader(r.myAddr)
		http.NewRequest(http.MethodPost, r.registryAddr, body)
		time.Sleep(r.interval)
	}
}
