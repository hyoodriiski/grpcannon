// Package cooldown implements a per-key cooldown gate for rate-limiting
// repeated actions within a configurable time window.
//
// A cooldown gate allows an action to proceed only if enough time has elapsed
// since the last time the action was performed for a given key. This is useful
// for suppressing repeated events such as error notifications, retries, or
// log messages that should not fire more than once per interval.
//
// Basic usage:
//
//	gate := cooldown.NewGate(5 * time.Second)
//	if gate.Allow("my-key") {
//		// perform the rate-limited action
//	}
package cooldown
