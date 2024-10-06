package pool

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

var errEmptyClients = errors.New("empty hosts list to proxy request to")
var errNoClientsAvailable = errors.New("no available host to proxy request to")

type ForwarderProvider interface {
	Next() (Forwarder, error)
	Run(ctx context.Context)
}

type ClientRegistrar interface {
	RegisterClient(addr string)
	DeRegisterClient(addr string)
}

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

func NewPool(cfg *PoolConfig) (ForwarderProvider, ClientRegistrar) {
	p := &ForwarderPool{
		maxAgeNoNotif: cfg.MaxAgeNoNotif,
		lastEntryIdx:  0,
		entries:       []Forwarder{},
		notifTimes:    map[string]time.Time{},
	}
	return p, p
}

func (cp *ForwarderPool) Next() (Forwarder, error) {
	cp.lock.Lock()
	defer cp.lock.Unlock()

	if len(cp.entries) == 0 {
		return nil, errEmptyClients
	}

	if cp.lastEntryIdx >= len(cp.entries) {
		cp.lastEntryIdx = 0
	}

	idx := cp.lastEntryIdx

	var hostEntry Forwarder
	found := false
	// check whether this Forwarder can actually handle the call; if not, try the next one.
	// If you went through the complete list and haven't found anything, return error.
	for {
		hostEntry = cp.entries[idx]
		idx++
		if hostEntry.CanForward() {
			found = true
			break
		}

		if idx >= len(cp.entries) {
			idx = 0
		}

		if idx == cp.lastEntryIdx {
			break
		}
	}
	if !found {
		return nil, errNoClientsAvailable
	}

	cp.lastEntryIdx = idx
	return hostEntry, nil
}

func (cp *ForwarderPool) RegisterClient(addr string) {
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

func (cp *ForwarderPool) DeRegisterClient(addr string) {
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
