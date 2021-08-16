package cache

import (
	"github.com/Bose/minisentinel"
	"github.com/alicebob/miniredis/v2"
)

// CreateRedisServer generates an in memory redis instance with sentinels
func CreateRedisServer() (*miniredis.Miniredis, *minisentinel.Sentinel) {
	m, _ := miniredis.Run()
	s := minisentinel.NewSentinel(m, minisentinel.WithReplica(m))
	s.Start()

	return m, s
}
