package pool

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var errNoClients = errors.New("no available hosts to proxy request to")

type Config struct {
	MaxAgeNoNotif time.Duration
	ListenAddr    string
}

type ClientPool struct {
	registerListenAddr string
	lock               sync.Mutex
	maxAgeNoNotif      time.Duration
	lastAddrIdx        int
	addrs              []string
	notifTimes         map[string]time.Time
	mux                *http.ServeMux
}

func New(cfg *Config) *ClientPool {
	cp := &ClientPool{
		registerListenAddr: cfg.ListenAddr,
		maxAgeNoNotif:      cfg.MaxAgeNoNotif,
		lastAddrIdx:        0,
		addrs:              []string{},
		notifTimes:         map[string]time.Time{},
		mux:                http.NewServeMux(),
	}

	cp.mux.HandleFunc(fmt.Sprintf("%s /", http.MethodPost), cp.registerClient)
	return cp
}

func (cp *ClientPool) Next() (string, error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	if len(cp.addrs) == 0 {
		return "", errNoClients
	}

	if cp.lastAddrIdx >= len(cp.addrs) {
		cp.lastAddrIdx = 0
	}

	addr := cp.addrs[cp.lastAddrIdx]
	cp.lastAddrIdx++
	return addr, nil
}

func (cp *ClientPool) ListenForClients(ctx context.Context) error {
	go cp.runCleanPool()

	server := &http.Server{Addr: cp.registerListenAddr, Handler: cp.mux}

	// listen for context to stop server gracefully
	go func() {
		<-ctx.Done()
		log.Printf("Gracefully sutting down client listener...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Fatalf("Server Shutdown Failed: %v", err)
		}
	}()

	return server.ListenAndServe()
}

func (cp *ClientPool) registerClient(w http.ResponseWriter, req *http.Request) {
	bytes, err := io.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	addr := string(bytes)
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if _, ok := cp.notifTimes[addr]; ok {
		// we already have this addr, update the time
		cp.notifTimes[addr] = time.Now()
		return
	}

	// this is a new client
	cp.addrs = append(cp.addrs, addr)
	cp.notifTimes[addr] = time.Now()
	log.Printf("INFO: added client %s", addr)
}

func (cp *ClientPool) runCleanPool() {
	t := time.NewTicker(time.Second)
	var needsClean bool
	for range t.C {
		needsClean = false
		for _, notifTime := range cp.notifTimes {
			if notifTime.Add(cp.maxAgeNoNotif).Before(time.Now()) {
				needsClean = true
				break
			}
		}

		if needsClean {
			cp.cleanPool()
		}
	}
}

func (cp *ClientPool) cleanPool() {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	var newAddrs []string
	newNotifTimes := map[string]time.Time{}
	for addr, notifTime := range cp.notifTimes {
		if notifTime.Add(cp.maxAgeNoNotif).Before(time.Now()) {
			log.Printf("INFO: removing client %s", addr)
			continue
		}
		newAddrs = append(newAddrs, addr)
		newNotifTimes[addr] = notifTime
	}
	cp.addrs = newAddrs
	cp.notifTimes = newNotifTimes
	cp.lastAddrIdx = 0
}
