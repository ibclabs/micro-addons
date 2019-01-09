package strategy

import (
	"sync/atomic"

	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/selector"
)

type roundRobin struct {
	tick int64
}

func NewRoundRobin() *roundRobin {
	return &roundRobin{}
}

func (r *roundRobin) Select(services []*registry.Service) selector.Next {
	var nodes []*registry.Node

	for _, service := range services {
		nodes = append(nodes, service.Nodes...)
	}

	return func() (*registry.Node, error) {
		if len(nodes) == 0 {
			return nil, selector.ErrNoneAvailable
		}

		node := nodes[int(atomic.AddInt64(&r.tick, 1))%len(nodes)]

		return node, nil
	}
}
