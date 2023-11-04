package grpc

import (
	"fmt"
	"time"

	"github.com/kovey/debug-go/debug"
	"github.com/kovey/discovery/etcd"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type Register struct {
	conf        etcd.Config
	cli         *etcd.Client
	shutdown    chan bool
	leaseId     clientv3.LeaseID
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse
	ins         *Instance
	ttl         int64
}

func NewRegister(conf etcd.Config) *Register {
	return &Register{conf: conf, shutdown: make(chan bool), cli: etcd.NewClient()}
}

func (r *Register) Register(ins *Instance, ttl int64) error {
	if !ins.Valid() {
		return fmt.Errorf("ddr[%s] has invalid ip", ins.Addr)
	}

	if err := r.cli.Connect(r.conf); err != nil {
		return err
	}

	r.ins = ins
	r.ttl = ttl
	if err := r.register(); err != nil {
		return err
	}

	go r.keepAlive()
	return nil
}

func (r *Register) register() error {
	leaseId, err := r.cli.Grant(r.ttl)
	if err != nil {
		return err
	}

	r.leaseId = leaseId
	data, err := r.ins.Encode()
	if err != nil {
		return err
	}

	if err := r.cli.Put(r.ins.Path(), data, clientv3.WithLease(r.leaseId)); err != nil {
		return err
	}

	/**
	kac, err := r.cli.KeepAlive(r.leaseId)
	if err != nil {
		return err
	}

	r.keepAliveCh = kac
	*/
	return nil
}

func (r *Register) unregister() error {
	return r.cli.Delete(r.ins.Path(), false)
}

func (r *Register) keepAlive() {
	ticker := time.NewTicker(time.Duration(r.ttl-1) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.shutdown:
			if err := r.unregister(); err != nil {
				debug.Erro("Instance[%s] Addr[%s] unregister failure, error: %s", r.ins.Name, r.ins.Addr, err)
			}
			if err := r.cli.Revoke(r.leaseId); err != nil {
				debug.Erro("Instance[%s] Addr[%s] leaseId[%d] revoke failure, error: %s", r.ins.Name, r.ins.Addr, r.leaseId, err)
			}

			r.cli.Shutdown()
			return
		case <-ticker.C:
			/**
			if r.keepAliveCh == nil {
				continue
			}
			*/

			resp, err := r.cli.KeepAliveOnce(r.leaseId)
			if err != nil {
				debug.Erro("keep alive failure, error: %s", err)
				if err := r.register(); err != nil {
					debug.Erro("Instance[%s] Addr[%s] register failure, error: %s", r.ins.Name, r.ins.Addr, err)
				}
				continue
			}

			debug.Info("keep alive success, lease: %d, TTL: %d", resp.ID, resp.TTL)
			///case res, ok := <-r.keepAliveCh:
			//debug.Info("keep alive res: %v, %t, chan: %+v", res, ok, r.keepAliveCh)
			/**
			if res == nil {
				if err := r.register(); err != nil {
					debug.Erro("Instance[%s] Addr[%s] register failure, error: %s", r.ins.Name, r.ins.Addr, err)
				}
			}
			*/
		}
	}
}

func (r *Register) GetInstance() (*Instance, error) {
	values, err := r.cli.Get(r.ins.Path(), false)
	if err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, fmt.Errorf("instance[%s] is empty", r.ins.Path())
	}

	ins := &Instance{}
	if err := ins.Decode(values.Get(r.ins.Path())); err != nil {
		return nil, err
	}

	return ins, nil
}

func (r *Register) Shutdown() {
	r.shutdown <- true
}
