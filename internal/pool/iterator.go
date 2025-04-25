package pool

import "sync"

type serverIterator struct {
	servers []*Server
	serverx int
	m       sync.Mutex
}

func newServerIterator(servers []*Server) *serverIterator {
	return &serverIterator{
		servers: servers,
		serverx: 0,
	}
}

func (si *serverIterator) Next() (*Server, bool) {
	si.m.Lock()
	defer si.m.Unlock()
	var server *Server

	for server == nil {
		s := si.servers[si.serverx]
		if s.Alive() {
			server = s
		}

		si.Inc()
	}

	return server, true
}

func (si *serverIterator) Inc() {
	si.serverx++
	if si.serverx == len(si.servers) {
		si.serverx = 0
	}
}
