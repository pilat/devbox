package depgraph

import (
	"context"
	"sync"
)

func Exec[T Entity](ctx context.Context, rounds [][]T, fn func(context.Context, T) error) error {
	return exec(ctx, rounds, false, fn)
}

func ExecReverse[T Entity](ctx context.Context, rounds [][]T, fn func(context.Context, T) error) error {
	return exec(ctx, rounds, true, fn)
}

func exec[T Entity](ctx context.Context, rounds [][]T, reverse bool, fn func(context.Context, T) error) error {
	if reverse {
		for i := len(rounds) - 1; i >= 0; i-- {
			if err := execStep(ctx, rounds[i], fn); err != nil {
				return err
			}
		}
	}

	for _, round := range rounds {
		if err := execStep(ctx, round, fn); err != nil {
			return err
		}
	}

	return nil
}

func execStep[T Entity](ctx context.Context, step []T, fn func(context.Context, T) error) error {
	wg := sync.WaitGroup{}
	mu := sync.Mutex{}
	var outErr error

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, entity := range step {
		wg.Add(1)
		go func(i int, entity T) {
			defer wg.Done() // Signal completion of this goroutine

			err := fn(ctx, entity)
			if err != nil {
				mu.Lock()
				if outErr == nil {
					outErr = err
				}
				mu.Unlock()

				cancel()
			}
		}(i, entity)
	}

	wg.Wait()

	return outErr
}
