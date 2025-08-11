package httpadp

import (
    "sync"
    "github.com/robjsliwa/pulse/app"
    "github.com/robjsliwa/pulse/domain"
)

// Broadcaster is a simple in-memory results broadcaster.
type Broadcaster struct {
    mu      sync.RWMutex
    subs    map[uint]map[chan domain.Results]struct{}
}

func NewBroadcaster() *Broadcaster { return &Broadcaster{subs: make(map[uint]map[chan domain.Results]struct{})} }

var _ app.ResultsStreamer = (*Broadcaster)(nil)

func (b *Broadcaster) Broadcast(pollID uint, res domain.Results) {
    b.mu.RLock()
    defer b.mu.RUnlock()
    for ch := range b.subs[pollID] {
        select { case ch <- res: default: }
    }
}

func (b *Broadcaster) Subscribe(pollID uint) (<-chan domain.Results, func()) {
    ch := make(chan domain.Results, 8)
    b.mu.Lock()
    if _, ok := b.subs[pollID]; !ok { b.subs[pollID] = make(map[chan domain.Results]struct{}) }
    b.subs[pollID][ch] = struct{}{}
    b.mu.Unlock()
    cancel := func() {
        b.mu.Lock()
        if m, ok := b.subs[pollID]; ok {
            delete(m, ch)
            if len(m) == 0 { delete(b.subs, pollID) }
        }
        b.mu.Unlock()
        close(ch)
    }
    return ch, cancel
}

