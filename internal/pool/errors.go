package pool

import (
	"errors"
)

var (
	errNoServersAvailable = errors.New("no servers available")
)
