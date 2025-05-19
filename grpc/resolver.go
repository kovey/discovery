package grpc

import (
	"sync"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/discovery/etcd"
	"google.golang.org/grpc/resolver"
)

type Resolver struct {
	cli    *etcd.Client
	prefix bool
	conn   resolver.ClientConn
	key    string
	lock   sync.Mutex
}

func NewResolver(conn resolver.ClientConn, cli *etcd.Client, key string, prefix bool) *Resolver {
	return &Resolver{conn: conn, cli: cli, prefix: prefix, key: key}
}

func (r *Resolver) Modify(key, value string) error {
	return r.doResolver()
}

func (r *Resolver) Create(key, value string) error {
	return r.doResolver()
}

func (r *Resolver) Delete(key, value string) error {
	return r.doResolver()
}

func (r *Resolver) Key() string {
	return r.key
}

func (r *Resolver) IsPrefix() bool {
	return r.prefix
}

func (r *Resolver) Tick(t time.Time) error {
	return r.doResolver()
}

func (r *Resolver) doResolver() error {
	values, err := r.cli.Get(r.key, r.prefix)
	if err != nil {
		return err
	}

	state := resolver.State{}
	for _, value := range values {
		ins := &Instance{}
		if err := ins.Decode(value); err != nil {
			debug.Erro("Decode Instance failure, error: %s", err)
			continue
		}

		state.Addresses = append(state.Addresses, ins.Address())
	}

	return r.update(state)
}

func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {
	if err := r.doResolver(); err != nil {
		debug.Erro("resolver service[%s] failure: %s", r.key, err)
	}
}

func (r *Resolver) Close() {
	r.cli.Close(r.key)
}

func (r *Resolver) update(state resolver.State) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.conn.UpdateState(state)
}

func (r *Resolver) start() error {
	if err := r.cli.Watch(r); err != nil {
		return err
	}

	return r.doResolver()
}
