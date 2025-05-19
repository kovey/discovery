package grpc

import (
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/discovery/etcd"
	"google.golang.org/grpc/resolver"
)

type DeleteInterface interface {
	Delete(serviceName string)
}

type Resolver struct {
	cli      *etcd.Client
	prefix   bool
	servList []resolver.Address
	conn     resolver.ClientConn
	key      string
	ev       DeleteInterface
}

func NewResolver(conn resolver.ClientConn, cli *etcd.Client, key string, prefix bool) *Resolver {
	return &Resolver{conn: conn, cli: cli, prefix: prefix, key: key}
}

func (r *Resolver) Event(ev DeleteInterface) {
	r.ev = ev
}

func (r *Resolver) Modify(key, value string) error {
	return r.add(value)
}

func (r *Resolver) Create(key, value string) error {
	return r.add(value)
}

func (r *Resolver) Delete(key, value string) error {
	ins := &Instance{}
	var tmp = -1
	for i, info := range r.servList {
		ins.Parse(info)
		if ins.Path() == key {
			tmp = i
			break
		}
	}

	if tmp < 0 {
		return nil
	}

	r.servList = append(r.servList[:tmp], r.servList[tmp+1:]...)
	return r.update()
}

func (r *Resolver) Key() string {
	return r.key
}

func (r *Resolver) IsPrefix() bool {
	return r.prefix
}

func (r *Resolver) Tick(t time.Time) error {
	return r.sync()
}

func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {
}

func (r *Resolver) Close() {
	r.cli.Close(r.key)
}

func (r *Resolver) update() error {
	return r.conn.UpdateState(resolver.State{Addresses: r.servList})
}

func (r *Resolver) sync() error {
	values, err := r.cli.Get(r.key, r.prefix)
	if err != nil {
		return err
	}

	r.servList = make([]resolver.Address, 0)
	for _, value := range values {
		ins := &Instance{}
		if err := ins.Decode(value); err != nil {
			debug.Erro("Decode Instance failure, error: %s", err)
		}

		r.servList = append(r.servList, ins.Address())
	}

	if len(r.servList) == 0 && r.ev != nil {
		r.ev.Delete(r.key)
	}

	return r.update()
}

func (r *Resolver) start() error {
	if err := r.cli.Watch(r); err != nil {
		return err
	}

	return r.sync()
}

func (r *Resolver) add(value string) error {
	ins := &Instance{}
	if err := ins.Decode(value); err != nil {
		return err
	}

	addr := ins.Address()
	for _, old := range r.servList {
		if old.Addr == addr.Addr {
			return nil
		}
	}

	r.servList = append(r.servList, addr)
	return r.update()
}
