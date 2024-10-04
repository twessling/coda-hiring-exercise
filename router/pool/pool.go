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
	lastEntryIdx  int
	entries       []*poolEntry
	notifTimes    map[string]time.Time
}

func NewPool(cfg *PoolConfig) *ClientPool {
	return &ClientPool{
		maxAgeNoNotif: cfg.MaxAgeNoNotif,
		lastEntryIdx:  0,
		entries:       []*poolEntry{},
		notifTimes:    map[string]time.Time{},
	}
}

func (cp *ClientPool) Next() (Forwarder, error) {
	return cp.next()
}

func (cp *ClientPool) next() (*poolEntry, error) {
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

func (cp *ClientPool) registerClient(addr string) {
	cp.lock.Lock()
	defer cp.lock.Unlock()
	if _, ok := cp.notifTimes[addr]; ok {
		// we already have this addr, only update the last notif time
		cp.notifTimes[addr] = time.Now()
		return
	}

	// this is a new client
	cp.entries = append(cp.entries, &poolEntry{Addr: addr, weight: 100, movingWindow: newMovingWindow()})
	cp.notifTimes[addr] = time.Now()
	log.Printf("INFO: added client %s for a total of %d", addr, len(cp.entries))
}

func (cp *ClientPool) deRegisterClient(addr string) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	delete(cp.notifTimes, addr)
	for i := 0; i < len(cp.entries); i++ {
		if cp.entries[i].Addr == addr {
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

	newHostEntries := []*poolEntry{}
	newNotifTimes := map[string]time.Time{}
	var removed []string
	for _, hostEntry := range cp.entries {
		addr := hostEntry.Addr
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
