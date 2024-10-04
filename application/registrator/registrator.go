package registrator

import (
	"context"
	"log"
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

func (r *Registrator) Run(ctx context.Context) error {
	t := time.NewTicker(r.interval)
	defer r.deregister()
	for {
		select {
		case <-ctx.Done():
			log.Print("INFO: Gracefully shutting down registrator..")
			return nil
		case <-t.C:
			body := strings.NewReader(r.myAddr)
			req, err := http.NewRequest(http.MethodPost, r.registryAddr, body)
			if err != nil {
				return err
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Printf("ERROR: while calling registrator: %v", err)
				// don't want to die here
				continue
			}

			if resp.StatusCode < 200 || resp.StatusCode >= 300 {
				log.Printf("WARN: registrator returned non-200: %d", resp.StatusCode)
			}
		}
	}
}

func (r *Registrator) deregister() {
	body := strings.NewReader(r.myAddr)
	req, err := http.NewRequest(http.MethodDelete, r.registryAddr, body)
	if err != nil {
		// we're shutting down anyway, best effort here
		return
	}
	_, _ = http.DefaultClient.Do(req)
	// we're shutting down anyway, best effort here.. not much to do with error or response code at this point.
}
