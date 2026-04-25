// Package jitter adds randomised delay to retry/backoff strategies
// to avoid thundering-herd problems under load.
package jitter

import (
	"math/rand"
	"time"
)

// Strategy controls how jitter is applied to a base duration.
type Strategy int

const (
	// Full replaces the base duration with a random value in [0, base].
	Full Strategy = iota
	// Equal adds a random value in [0, base/2] to base/2.
	Equal
	// Decorrelated picks a random value in [base, prev*3], anchored to the
	// previous delay (pass 0 for the first call).
	Decorelated
)

// Jitter holds configuration for applying jitter.
type Jitter struct {
	strategy Strategy
	rng      *rand.Rand
}

// New returns a Jitter using the given strategy.
// A nil source uses the default global random source.
func New(s Strategy) *Jitter {
	return &Jitter{
		strategy: s,
		rng:      rand.New(rand.NewSource(time.Now().UnixNano())), //nolint:gosec
	}
}

// Apply returns a jittered duration derived from base.
// prev is only used by the Decorrelated strategy; pass 0 otherwise.
func (j *Jitter) Apply(base, prev time.Duration) time.Duration {
	if base <= 0 {
		return 0
	}
	switch j.strategy {
	case Full:
		return time.Duration(j.rng.Int63n(int64(base) + 1))
	case Equal:
		half := base / 2
		return half + time.Duration(j.rng.Int63n(int64(half)+1))
	case Decorrelated:
		if prev <= 0 {
			prev = base
		}
		lo := int64(base)
		hi := int64(prev) * 3
		if hi <= lo {
			return base
		}
		return time.Duration(lo + j.rng.Int63n(hi-lo))
	default:
		return base
	}
}

// Capped returns a jittered duration derived from base, clamped to a maximum
// value of cap. This is useful when combining jitter with exponential backoff
// to prevent delays from growing without bound.
func (j *Jitter) Capped(base, prev, cap time.Duration) time.Duration {
	if cap <= 0 {
		return j.Apply(base, prev)
	}
	d := j.Apply(base, prev)
	if d > cap {
		return cap
	}
	return d
}
