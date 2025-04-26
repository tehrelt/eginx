package pool

import "sync"

type serverIterator struct {
	buf []*Server
	x   int
	m   sync.Mutex
}

func newServerIterator(servers []*Server) *serverIterator {
	return &serverIterator{
		buf: servers,
		x:   0,
	}
}

func (si *serverIterator) Next() (*Server, bool) {
	si.m.Lock()
	defer si.m.Unlock()
	var server *Server

	for server == nil {
		s := si.buf[si.x]
		if s.Alive() {
			server = s
		}

		si.Inc()
	}

	return server, true
}

func (si *serverIterator) Inc() {
	si.x++
	if si.x == len(si.buf) {
		si.x = 0
	}
}

func (si *serverIterator) Set(s []*Server) {
	si.m.Lock()
	defer si.m.Unlock()

	for _, srv := range si.buf {
		srv.health.Stop()
	}

	si.buf = s
}
