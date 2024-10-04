package pool

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var errNoClients = errors.New("no available hosts to proxy request to")

type PoolConfig struct {
	MaxAgeNoNotif time.Duration
}

type ClientPool struct {
	lock          sync.Mutex
	maxAgeNoNotif time.Duration
	lastAddrIdx   int
	addrs         []string
	notifTimes    map[string]time.Time
}

func NewPool(cfg *PoolConfig) *ClientPool {
	cp := &ClientPool{
		maxAgeNoNotif: cfg.MaxAgeNoNotif,
		lastAddrIdx:   0,
		addrs:         []string{},
		notifTimes:    map[string]time.Time{},
	}

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

func (cp *ClientPool) registerClient(addr string) {
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

func (cp *ClientPool) Run(ctx context.Context) {
	t := time.NewTicker(time.Second)
	var needsClean bool

	for {
		select {
		case <-t.C:
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
		case <-ctx.Done():
			return
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
