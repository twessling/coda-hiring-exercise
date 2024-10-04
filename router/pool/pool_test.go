package pool

import (
	"reflect"
	"testing"
	"time"
)

func TestEmptyPool(t *testing.T) {

	pool := &ClientPool{
		maxAgeNoNotif: time.Hour,
		lastAddrIdx:   0,
		addrs:         []string{},
		notifTimes:    map[string]time.Time{},
	}

	if val, err := pool.Next(); err == nil {
		t.Fatalf("expected error but got %v", val)
	}
}

func TestPoolNext(t *testing.T) {

	tests := map[string]struct {
		addrs           []string
		wantNext10Times []string
		startIndex      int
	}{
		"1 addr": {
			addrs:           []string{"purple"},
			wantNext10Times: repeat("purple", 10),
		},
		"3 addrs": {
			addrs:           []string{"purple", "green", "yellow"},
			wantNext10Times: []string{"purple", "green", "yellow", "purple", "green", "yellow", "purple", "green", "yellow", "purple"},
		},
		"12 addrs": {
			addrs: []string{
				"purple",
				"green",
				"yellow",
				"blue",
				"magenta",
				"white",
				"black",
				"red",
				"brown",
				"cyan",
				"grey",
				"pink",
			},
			wantNext10Times: []string{"purple", "green", "yellow", "blue", "magenta", "white", "black", "red", "brown", "cyan"},
		},
		"3 addrs with invalid startindex": {
			addrs:           []string{"purple", "green", "yellow"},
			wantNext10Times: []string{"purple", "green", "yellow", "purple", "green", "yellow", "purple", "green", "yellow", "purple"},
			startIndex:      10,
		},
		"3 addrs with non-0 startindex": {
			addrs:           []string{"purple", "green", "yellow"},
			wantNext10Times: []string{"yellow", "purple", "green", "yellow", "purple", "green", "yellow", "purple", "green", "yellow"},
			startIndex:      2,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			notifTimes := func(entries []string) map[string]time.Time {
				m := map[string]time.Time{}
				for _, e := range entries {
					m[e] = time.Now()
				}
				return m
			}

			pool := &ClientPool{
				maxAgeNoNotif: time.Hour, // not used in this test anyway
				lastAddrIdx:   test.startIndex,
				addrs:         test.addrs,
				notifTimes:    notifTimes(test.addrs),
			}

			var res []string
			for i := 0; i < 10; i++ {
				n, err := pool.Next()
				if err != nil {
					t.Fatalf("got unexpected error: %v", err)
				}
				res = append(res, n)
			}
			if !reflect.DeepEqual(res, test.wantNext10Times) {
				t.Fatalf("list differs: got %v want %v", res, test.wantNext10Times)
			}
		})
	}
}

func repeat(val string, n int) []string {
	var res []string
	for i := 0; i < n; i++ {
		res = append(res, val)
	}
	return res
}

func TestPoolRegisterClient(t *testing.T) {
	tests := map[string]struct {
		addrsToRegister []string
		resultingAddrs  []string
	}{
		"no addresses": {
			addrsToRegister: []string{},
			resultingAddrs:  []string{},
		},
		"one address": {
			addrsToRegister: []string{"somewhere.org"},
			resultingAddrs:  []string{"somewhere.org"},
		},
		"several addresses": {
			addrsToRegister: []string{"somewhere.org", "there.com", "thisplace.gov"},
			resultingAddrs:  []string{"somewhere.org", "there.com", "thisplace.gov"},
		},
		"several addresses with overlap": {
			addrsToRegister: []string{"somewhere.org", "there.com", "thisplace.gov", "somewhere.org", "there.com", "somewhere.org"},
			resultingAddrs:  []string{"somewhere.org", "there.com", "thisplace.gov"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pool := &ClientPool{
				maxAgeNoNotif: time.Hour,
				lastAddrIdx:   0,
				addrs:         []string{},
				notifTimes:    map[string]time.Time{},
			}
			for _, addr := range test.addrsToRegister {
				pool.registerClient(addr)
			}

			if !reflect.DeepEqual(pool.addrs, test.resultingAddrs) {
				t.Fatalf("difference in addrs: got %v want %v", pool.addrs, test.resultingAddrs)
			}
		})
	}
}

func TestCleanPool(t *testing.T) {
	type hostEntry struct {
		addr      string
		lastNotif time.Time
	}

	tests := map[string]struct {
		hosts           []*hostEntry
		maxAgeNoNotif   time.Duration
		addrsAfterClean []string
	}{
		"no addresses": {
			hosts:           []*hostEntry{},
			addrsAfterClean: []string{},
			maxAgeNoNotif:   time.Second,
		},
		"one address, not cleaned": {
			hosts: []*hostEntry{
				{addr: "there.com", lastNotif: time.Now()},
			},
			maxAgeNoNotif:   time.Second,
			addrsAfterClean: []string{"there.com"},
		},
		"one address, cleaned": {
			hosts: []*hostEntry{
				{addr: "there.com", lastNotif: time.Now().Add(-time.Hour)}, // hour old
			},
			maxAgeNoNotif:   time.Second,
			addrsAfterClean: []string{},
		},
		"multiple addresses, some cleaned": {
			hosts: []*hostEntry{
				{addr: "there.com", lastNotif: time.Now().Add(-time.Hour)}, // hour old
				{addr: "here.com", lastNotif: time.Now()},
				{addr: "onthemoon.com", lastNotif: time.Now()},
				{addr: "myplace.com", lastNotif: time.Now().Add(-5 * time.Minute)}, // 5 mins old
				{addr: "yourplace.com", lastNotif: time.Now()},
			},
			maxAgeNoNotif:   time.Second,
			addrsAfterClean: []string{"here.com", "onthemoon.com", "yourplace.com"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {

			addrs, notifTimes := func(hosts []*hostEntry) ([]string, map[string]time.Time) {
				a := []string{}
				m := map[string]time.Time{}
				for _, e := range hosts {
					a = append(a, e.addr)
					m[e.addr] = e.lastNotif
				}
				return a, m
			}(test.hosts)

			pool := &ClientPool{
				maxAgeNoNotif: test.maxAgeNoNotif,
				addrs:         addrs,
				notifTimes:    notifTimes,
			}

			pool.cleanPool()

			if !reflect.DeepEqual(pool.addrs, test.addrsAfterClean) {
				t.Fatalf("difference in addrs: got %v want %v", pool.addrs, test.addrsAfterClean)
			}
		})
	}
}

func TestNextAndCleanInteraction(t *testing.T) {
	type hostEntry struct {
		addr      string
		lastNotif time.Time
	}

	tests := map[string]struct {
		hosts                      []*hostEntry
		maxAgeNoNotif              time.Duration
		wantNext10TimesBeforeClean []string
		wantNext10TimesAfterClean  []string
	}{
		"one address, not cleaned": {
			hosts: []*hostEntry{
				{addr: "there.com", lastNotif: time.Now()},
			},
			maxAgeNoNotif:              time.Second,
			wantNext10TimesBeforeClean: repeat("there.com", 10),
			wantNext10TimesAfterClean:  repeat("there.com", 10),
		},
		"multiple addresses, some cleaned": {
			hosts: []*hostEntry{
				{addr: "there.com", lastNotif: time.Now().Add(-time.Hour)}, // hour old
				{addr: "here.com", lastNotif: time.Now()},
				{addr: "onthemoon.com", lastNotif: time.Now()},
				{addr: "myplace.com", lastNotif: time.Now().Add(-5 * time.Minute)}, // 5 mins old
				{addr: "yourplace.com", lastNotif: time.Now()},
			},
			maxAgeNoNotif:              time.Second,
			wantNext10TimesBeforeClean: []string{"there.com", "here.com", "onthemoon.com", "myplace.com", "yourplace.com", "there.com", "here.com", "onthemoon.com", "myplace.com", "yourplace.com"},
			wantNext10TimesAfterClean:  []string{"here.com", "onthemoon.com", "yourplace.com", "here.com", "onthemoon.com", "yourplace.com", "here.com", "onthemoon.com", "yourplace.com", "here.com"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			addrs, notifTimes := func(hosts []*hostEntry) ([]string, map[string]time.Time) {
				a := []string{}
				m := map[string]time.Time{}
				for _, e := range hosts {
					a = append(a, e.addr)
					m[e.addr] = e.lastNotif
				}
				return a, m
			}(test.hosts)

			pool := &ClientPool{
				maxAgeNoNotif: test.maxAgeNoNotif,
				addrs:         addrs,
				notifTimes:    notifTimes,
			}

			var beforeAddrs []string
			for i := 0; i < 10; i++ {
				a, err := pool.Next()
				if err != nil {
					t.Fatalf("unexpected error while Next: %v", err)
				}
				beforeAddrs = append(beforeAddrs, a)
			}
			if !reflect.DeepEqual(beforeAddrs, test.wantNext10TimesBeforeClean) {
				t.Fatalf("difference in addrs: got %+v want %+v", beforeAddrs, test.wantNext10TimesBeforeClean)
			}

			pool.cleanPool()

			var afterAddrs []string
			for i := 0; i < 10; i++ {
				a, err := pool.Next()
				if err != nil {
					t.Fatalf("unexpected error while Next: %v", err)
				}
				afterAddrs = append(afterAddrs, a)
			}

			if !reflect.DeepEqual(afterAddrs, test.wantNext10TimesAfterClean) {
				t.Fatalf("difference in addrs: got %v want %v", afterAddrs, test.wantNext10TimesAfterClean)
			}
		})
	}
}

func TestDeregisterClient(t *testing.T) {
	tests := map[string]struct {
		addrsToRegister   []string
		addrsToDeregister []string
		resultingAddrs    []string
	}{
		"no addresses": {
			addrsToRegister:   []string{},
			addrsToDeregister: []string{},
			resultingAddrs:    []string{},
		},
		"one address": {
			addrsToRegister:   []string{"somewhere.org"},
			addrsToDeregister: []string{"somewhere.org"},
			resultingAddrs:    []string{},
		},
		"several addresses": {
			addrsToRegister:   []string{"somewhere.org", "there.com", "thisplace.gov"},
			addrsToDeregister: []string{"there.com"},
			resultingAddrs:    []string{"somewhere.org", "thisplace.gov"},
		},
		"unknown address to deregister": {
			addrsToRegister:   []string{"somewhere.org", "thisplace.gov"},
			addrsToDeregister: []string{"there.com"},
			resultingAddrs:    []string{"somewhere.org", "thisplace.gov"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			pool := &ClientPool{
				maxAgeNoNotif: time.Hour,
				lastAddrIdx:   0,
				addrs:         []string{},
				notifTimes:    map[string]time.Time{},
			}
			for _, addr := range test.addrsToRegister {
				pool.registerClient(addr)
			}

			for _, addr := range test.addrsToDeregister {
				pool.deRegisterClient(addr)
			}

			if !reflect.DeepEqual(pool.addrs, test.resultingAddrs) {
				t.Fatalf("difference in addrs: got %v want %v", pool.addrs, test.resultingAddrs)
			}
		})
	}
}
