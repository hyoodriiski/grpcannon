package budget

// InvokerFn is a function that performs a single gRPC invocation.
type InvokerFn func() error

// Guard wraps an InvokerFn with budget enforcement. If the budget is
// already exhausted before the call, ErrExhausted is returned immediately
// without invoking next. After the call, the outcome is recorded.
func Guard(b *Budget, next InvokerFn) error {
	if b.Exhausted() {
		return ErrExhausted
	}
	err := next()
	if recordErr := b.Record(err != nil); recordErr != nil {
		// Budget just became exhausted; surface budget error, not call error.
		return recordErr
	}
	return err
}
