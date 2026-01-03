package lb

import (
	"sync"
	"time"

	"github.com/nerdneilsfield/RSSHub-Gateway/internal/upstream"
)

type WRRPicker struct {
	items []*upstream.State
	curr  []int
	mu    sync.Mutex
}

func NewWRRPicker(items []*upstream.State) *WRRPicker {
	return &WRRPicker{
		items: items,
		curr:  make([]int, len(items)),
	}
}

func (p *WRRPicker) Pick(_ string, now time.Time, avoid map[*upstream.State]struct{}) *upstream.State {
	p.mu.Lock()
	defer p.mu.Unlock()
	var best *upstream.State
	bestIdx := -1
	totalWeight := 0
	for i, item := range p.items {
		if avoid != nil {
			if _, skip := avoid[item]; skip {
				continue
			}
		}
		if !item.Available(now) {
			continue
		}
		p.curr[i] += item.Weight
		totalWeight += item.Weight
		if best == nil || p.curr[i] > p.curr[bestIdx] {
			best = item
			bestIdx = i
		}
	}
	if best == nil {
		return nil
	}
	p.curr[bestIdx] -= totalWeight
	return best
}
