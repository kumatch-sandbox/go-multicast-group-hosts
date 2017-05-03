package main

import (
	"log"
	"sync"
)

const lifeTime = 100

type host struct {
	addr string
	life int
}

func newHost(addr string) *host {
	return &host{
		addr: addr,
	}
}

func (h *host) refresh() {
	h.life = lifeTime
}

func (h *host) check() bool {
	h.life--

	if h.life < 1 {
		return false
	}

	return true
}

type hosts struct {
	entries map[string]*host
	updated chan struct{}
	mu      sync.RWMutex
}

func newHosts() *hosts {
	return &hosts{
		entries: make(map[string]*host),
		updated: make(chan struct{}, 256),
	}
}

func (h *hosts) Add(addr string) {
	h.mu.Lock()

	if host, ok := h.entries[addr]; ok {
		host.refresh()
	} else {
		host := newHost(addr)
		host.refresh()
		h.entries[addr] = host
		h.updated <- struct{}{}
	}

	h.mu.Unlock()
}

func (h *hosts) Check() {
	h.mu.Lock()
	var updated bool

	for _, host := range h.entries {
		if !host.check() {
			log.Printf("updated: %v\n", host)
			delete(h.entries, host.addr)
			updated = true
		}
	}

	if updated {
		h.updated <- struct{}{}
	}

	h.mu.Unlock()
}

func (h *hosts) Display() {
	h.mu.RLock()
	log.Println("---------- Known hosts ----------")
	for addr := range h.entries {
		log.Println(addr)
	}
	log.Println("---------------------------------")
	h.mu.RUnlock()
}
