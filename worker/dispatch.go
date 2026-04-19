package worker

import "context"

// Dispatch sends n tasks into a channel and closes it.
// Each task is built by the provided factory function.
func Dispatch(ctx context.Context, n int, factory func(i int) Task) <-chan Task {
	ch := make(chan Task, 64)
	go func() {
		defer close(ch)
		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return
			case ch <- factory(i):
			}
		}
	}()
	return ch
}
