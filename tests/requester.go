package main

import (
	"encoding/json"
	"flag"
	"log/slog"
	"net/http"
	"sync"
)

var (
	count  int
	target string
)

func init() {
	flag.IntVar(&count, "count", 0, "")
	flag.StringVar(&target, "target", "", "")
}

func main() {
	flag.Parse()

	if count == 0 || target == "" {
		panic("invalid args")
	}

	counter := make(map[int]int)
	mutex := sync.Mutex{}
	wg := sync.WaitGroup{}

	workersCount := 25
	ch := make(chan struct{}, workersCount)

	for i := range workersCount {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			l := slog.With(slog.Int("worker", x))
			for range ch {
				l.Info("requesting", slog.String("target", target))

				response, err := http.Get(target)
				if err != nil {
					l.Error("failed request", slog.String("error", err.Error()))
					continue
				}

				res := struct {
					Server int `json:"server"`
				}{}

				if err := json.NewDecoder(response.Body).Decode(&res); err != nil {
					l.Error("failed decode", slog.String("error", err.Error()))
					continue
				}

				mutex.Lock()
				v, ok := counter[res.Server]
				if !ok {
					counter[res.Server] = 1
				} else {
					counter[res.Server] = v + 1
				}
				mutex.Unlock()
			}
		}(i)
	}

	for range count {
		ch <- struct{}{}
	}
	close(ch)

	wg.Wait()

	for key, value := range counter {
		slog.Info("result", slog.Int("server", key), slog.Int("count", value))
	}
}
