package workerpool

import (
	"context"
	"sync"
)

func Run[T any](ctx context.Context, count int, in <-chan T, fn func(T, int)) {
	wg := sync.WaitGroup{}
	wg.Add(count)
	workerx := 0

	for range count {
		go func(i int) {
			for {
				select {
				case <-ctx.Done():
					return
				case request, ok := <-in:
					if !ok {
						return
					}

					select {
					case <-ctx.Done():
						return
					default:
						fn(request, i)
					}
				}

			}
		}(workerx)
	}

	wg.Wait()
}
