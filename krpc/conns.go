package krpc

import (
	"fmt"
	"sync"

	"google.golang.org/grpc"
)

type conns struct {
	data   map[string]*grpc.ClientConn
	locker sync.RWMutex
}

func newConns() *conns {
	return &conns{data: make(map[string]*grpc.ClientConn), locker: sync.RWMutex{}}
}

func (c *conns) get(serviceName string) (*grpc.ClientConn, error) {
	c.locker.RLock()
	defer c.locker.RUnlock()

	if cc, ok := c.data[serviceName]; ok {
		return cc, nil
	}

	return nil, fmt.Errorf("service[%s] not found", serviceName)
}

func (c *conns) remove(serviceName string) {
	c.locker.Lock()
	defer c.locker.Unlock()
	delete(c.data, serviceName)
}

func (c *conns) add(serviceName string) (*grpc.ClientConn, error) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if cc, ok := c.data[serviceName]; ok {
		return cc, nil
	}

	cc, err := dial(serviceName)
	if err != nil {
		return cc, err
	}

	c.data[serviceName] = cc
	return cc, nil
}

func (c *conns) close() {
	c.locker.RLock()
	defer c.locker.RUnlock()
	for _, cc := range c.data {
		cc.Close()
	}
}

var c = newConns()

func Remove(serviceName string) {
	c.remove(serviceName)
}
