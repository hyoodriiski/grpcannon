// Package admit implements a lightweight admission controller for grpcannon.
// It limits the number of concurrent requests entering the load-test pipeline
// and immediately rejects excess requests with ErrRejected.
package admit
