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

type ForwarderPool struct {
	lock          sync.Mutex
	maxAgeNoNotif time.Duration
	lastEntryIdx  int
	entries       []Forwarder
	notifTimes    map[string]time.Time
}

func NewPool(cfg *PoolConfig) *ForwarderPool {
	return &ForwarderPool{
		maxAgeNoNotif: cfg.MaxAgeNoNotif,
		lastEntryIdx:  0,
		entries:       []Forwarder{},
		notifTimes:    map[string]time.Time{},
	}
}

func (cp *ForwarderPool) Next() (Forwarder, error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	if len(cp.entries) == 0 {
		return nil, errNoClients
	}

	if cp.lastEntryIdx >= len(cp.entries) {
		cp.lastEntryIdx = 0
	}

	hostEntry := cp.entries[cp.lastEntryIdx]
	cp.lastEntryIdx++
	return hostEntry, nil
}

func (cp *ForwarderPool) registerClient(addr string) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if _, ok := cp.notifTimes[addr]; ok {
		// we already have this addr, only update the last notif time
		cp.notifTimes[addr] = time.Now()
		return
	}

	// this is a new client
	cp.entries = append(cp.entries, newForwardHandler(addr))
	cp.notifTimes[addr] = time.Now()
	log.Printf("INFO: added client %s for a total of %d", addr, len(cp.entries))
}

func (cp *ForwarderPool) deRegisterClient(addr string) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	delete(cp.notifTimes, addr)
	for i := 0; i < len(cp.entries); i++ {
		if cp.entries[i].Host() == addr {
			if i == len(cp.entries)-1 {
				cp.entries = cp.entries[:i] // If it's the last element, return up to the last
			} else {
				cp.entries = append(cp.entries[:i], cp.entries[i+1:]...)
			}
			break
		}
	}
	if cp.lastEntryIdx >= len(cp.entries) {
		// reset when necessary. Next() checks for proper value, but better to be explicit.
		cp.lastEntryIdx = 0
	}
	log.Printf("INFO: deregistered client %s for a total of %d", addr, len(cp.entries))
}

func (cp *ForwarderPool) Run(ctx context.Context) {
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

func (cp *ForwarderPool) cleanPool() {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	newHostEntries := []Forwarder{}
	newNotifTimes := map[string]time.Time{}
	var removed []string
	for _, hostEntry := range cp.entries {
		addr := hostEntry.Host()
		notifTime := cp.notifTimes[addr]
		if notifTime.Add(cp.maxAgeNoNotif).Before(time.Now()) {
			removed = append(removed, addr)
			continue
		}
		newHostEntries = append(newHostEntries, hostEntry)
		newNotifTimes[addr] = notifTime
	}

	cp.entries = newHostEntries
	cp.notifTimes = newNotifTimes
	if cp.lastEntryIdx >= len(cp.entries) {
		// reset when necessary. Next() checks for proper value, but better to be explicit.
		cp.lastEntryIdx = 0
	}
	log.Printf("Pool cleanup done, removed %d items. New pool size: %d", len(removed), len(cp.entries))
}
