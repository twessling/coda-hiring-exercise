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
	return &ClientPool{
		maxAgeNoNotif: cfg.MaxAgeNoNotif,
		lastAddrIdx:   0,
		addrs:         []string{},
		notifTimes:    map[string]time.Time{},
	}
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
		// we already have this addr, only update the last notif time
		cp.notifTimes[addr] = time.Now()
		return
	}

	// this is a new client
	cp.addrs = append(cp.addrs, addr)
	cp.notifTimes[addr] = time.Now()
	log.Printf("INFO: added client %s for a total of %d", addr, len(cp.addrs))
}

func (cp *ClientPool) deRegisterClient(addr string) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	delete(cp.notifTimes, addr)
	for i := 0; i < len(cp.addrs); i++ {
		if cp.addrs[i] == addr {
			if i == len(cp.addrs)-1 {
				cp.addrs = cp.addrs[:i] // If it's the last element, return up to the last
			} else {
				cp.addrs = append(cp.addrs[:i], cp.addrs[i+1:]...)
			}
			break
		}
	}
	if cp.lastAddrIdx >= len(cp.addrs) {
		// reset when necessary. Next() checks for proper value, but better to be explicit.
		cp.lastAddrIdx = 0
	}
	log.Printf("INFO: deregistered client %s for a total of %d", addr, len(cp.addrs))
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
	newAddrs := []string{}
	newNotifTimes := map[string]time.Time{}
	var removed []string
	for _, addr := range cp.addrs {
		notifTime := cp.notifTimes[addr]
		if notifTime.Add(cp.maxAgeNoNotif).Before(time.Now()) {
			removed = append(removed, addr)
			continue
		}
		newAddrs = append(newAddrs, addr)
		newNotifTimes[addr] = notifTime
	}
	cp.addrs = newAddrs
	cp.notifTimes = newNotifTimes
	if cp.lastAddrIdx >= len(cp.addrs) {
		// reset when necessary. Next() checks for proper value, but better to be explicit.
		cp.lastAddrIdx = 0
	}
	log.Printf("Pool cleanup done, removed %d items. New pool size: %d", len(removed), len(cp.addrs))
}
