package pool

import (
	"log/slog"
	"sync"
)

type serverIterator struct {
	buf []*server
	x   int
	m   sync.Mutex
}

func newServerIterator(servers []*server) *serverIterator {
	return &serverIterator{
		buf: servers,
		x:   0,
	}
}

func (si *serverIterator) Next() (server *server, ok bool) {
	si.m.Lock()
	defer si.m.Unlock()

	start := si.x
	if len(si.buf) == 0 {
		return nil, false
	}

	for server == nil {
		s := si.buf[si.x]
		if s.Alive() {
			server = s
		} else {
			slog.Debug("server not alive", slog.String("addr", s.URL.String()))
		}

		si.Inc()
		if si.x == start && server == nil {
			slog.Debug("si.x equals start point and server still nil")
			return nil, false
		}
	}

	return server, true
}

func (si *serverIterator) Inc() {
	si.x++
	if si.x == len(si.buf) {
		si.x = 0
	}
}

func (si *serverIterator) Set(s []*server) {
	si.m.Lock()
	defer si.m.Unlock()

	for _, srv := range si.buf {
		srv.health.Stop()
	}

	si.buf = s
}
