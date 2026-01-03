package upstream

import (
	"net/url"
	"sync"
	"time"
)

type State struct {
	URL       *url.URL
	HostLabel string
	Weight    int
	AccessKey string

	mu               sync.Mutex
	healthy          bool
	healthFailCount  int
	consecutiveFails int
	ejectUntil       time.Time
	backoff          time.Duration
}

func NewState(parsed *url.URL, weight int, accessKey string) *State {
	hostLabel := parsed.Host
	return &State{
		URL:       parsed,
		HostLabel: hostLabel,
		Weight:    weight,
		AccessKey: accessKey,
		healthy:   true,
	}
}

func (s *State) Available(now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.healthy {
		return false
	}
	return now.After(s.ejectUntil) || now.Equal(s.ejectUntil)
}

func (s *State) SetHealthy(healthy bool) (changed bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	changed = s.healthy != healthy
	s.healthy = healthy
	if healthy {
		s.healthFailCount = 0
	}
	return changed
}

func (s *State) HealthFail(retries int) (becameUnhealthy bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.healthFailCount++
	if s.healthFailCount >= retries {
		if s.healthy {
			s.healthy = false
			return true
		}
	}
	return false
}

func (s *State) HealthSuccess() (becameHealthy bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.healthy {
		becameHealthy = true
	}
	s.healthy = true
	s.healthFailCount = 0
	return becameHealthy
}

func (s *State) RecordSuccess() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consecutiveFails = 0
}

func (s *State) RecordFailure(now time.Time, failThreshold int, baseEject time.Duration, maxEject time.Duration) (ejected bool, until time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.consecutiveFails++
	if s.consecutiveFails < failThreshold {
		return false, s.ejectUntil
	}
	s.consecutiveFails = 0
	if s.backoff == 0 {
		s.backoff = baseEject
	} else {
		s.backoff = s.backoff * 2
		if s.backoff > maxEject {
			s.backoff = maxEject
		}
	}
	s.ejectUntil = now.Add(s.backoff)
	return true, s.ejectUntil
}

func (s *State) Snapshot() (healthy bool, ejectUntil time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.healthy, s.ejectUntil
}
