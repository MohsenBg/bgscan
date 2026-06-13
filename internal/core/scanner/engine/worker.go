package engine

import (
	"context"
	"sync"
)

// runWorkerPool starts a fixed-size worker pool and blocks until all workers exit.
func runWorkerPool(
	ctx context.Context,
	workers int,
	pause *PauseController,
	input <-chan string,
	process func(string),
) {
	workers = getWorkerCount(workers)

	var wg sync.WaitGroup
	wg.Add(workers)

	for range workers {
		go func() {
			defer wg.Done()
			runWorker(ctx, pause, input, process)
		}()
	}

	wg.Wait()
}

// runWorker consumes items from the input channel until the context is cancelled
// or the channel is closed.
func runWorker(
	ctx context.Context,
	pause *PauseController,
	input <-chan string,
	process func(string),
) {
	for {
		if pause != nil && !pause.Wait(ctx) {
			return
		}

		select {
		case <-ctx.Done():
			return

		case item, ok := <-input:
			if !ok {
				return
			}

			process(item)
		}
	}
}

// getWorkerCount ensures that at least one worker is used.
func getWorkerCount(workers int) int {
	if workers <= 0 {
		return 1
	}

	return workers
}

