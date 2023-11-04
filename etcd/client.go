package etcd

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kovey/debug-go/debug"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type IEvent interface {
	Create(key, value string) error
	Modify(key, value string) error
	Delete(key, value string) error
	Key() string
	IsPrefix() bool
	Tick(t time.Time) error
}

type Client struct {
	cli     *clientv3.Client
	wait    sync.WaitGroup
	cancels map[string]func()
	locker  sync.RWMutex
}

func NewClient() *Client {
	return &Client{wait: sync.WaitGroup{}, cancels: make(map[string]func()), locker: sync.RWMutex{}}
}

func (c *Client) Connect(conf Config) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   conf.Endpoints,
		DialTimeout: time.Duration(conf.DialTimeout) * time.Second,
		Username:    conf.Username,
		Password:    conf.Password,
	})
	if err != nil {
		return err
	}

	c.cli = cli

	return nil
}

func (c *Client) Put(key, value string, opts ...clientv3.OpOption) error {
	ctx, cancel := context.WithTimeout(context.Background(), Req_Timeout*time.Second)
	defer cancel()
	kv := clientv3.NewKV(c.cli)
	if _, err := kv.Put(ctx, key, value, opts...); err != nil {
		return err
	}

	return nil
}

func (c *Client) Get(key string, isPrefix bool) (Values, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Req_Timeout*time.Second)
	defer cancel()
	kv := clientv3.NewKV(c.cli)
	var opts []clientv3.OpOption
	if isPrefix {
		opts = []clientv3.OpOption{clientv3.WithPrefix()}
	}

	data, err := kv.Get(ctx, key, opts...)
	if err != nil {
		return nil, err
	}

	res := make(Values)
	for _, dt := range data.Kvs {
		res.Add(string(dt.Key), string(dt.Value))
	}

	return res, nil
}

func (c *Client) Delete(key string, isPrefix bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), Req_Timeout*time.Second)
	defer cancel()
	kv := clientv3.NewKV(c.cli)
	var opts []clientv3.OpOption
	if isPrefix {
		opts = []clientv3.OpOption{clientv3.WithPrefix()}
	}

	_, err := kv.Delete(ctx, key, opts...)
	return err
}

func (c *Client) Watch(e IEvent, opts ...clientv3.OpOption) error {
	if e == nil {
		return errors.New("event is nil")
	}

	if e.IsPrefix() {
		opts = append(opts, clientv3.WithPrefix())
	}

	ctx, cancel := context.WithTimeout(context.Background(), Req_Timeout*time.Second)
	c.locker.Lock()
	c.cancels[e.Key()] = cancel
	c.locker.Unlock()
	c.wait.Add(1)
	go c.watch(ctx, e, opts)

	return nil
}

func (c *Client) watch(ctx context.Context, e IEvent, opts []clientv3.OpOption) {
	defer c.wait.Done()
	watcher := clientv3.NewWatcher(c.cli)
	defer watcher.Close()
	wch := watcher.Watch(ctx, e.Key(), opts...)
	ticker := time.NewTicker(time.Second * 60)
	defer ticker.Stop()

	for {
		select {
		case t, ok := <-ticker.C:
			if !ok {
				return
			}
			if err := e.Tick(t); err != nil {
				debug.Erro("etcd watch[%s] tick failure, error: %s", e.Key(), err)
			}
		case <-ctx.Done():
			return
		case data, ok := <-wch:
			if !ok {
				return
			}
			for _, ev := range data.Events {
				switch ev.Type {
				case mvccpb.PUT:
					if ev.IsCreate() {
						if err := e.Create(string(ev.Kv.Key), string(ev.Kv.Value)); err != nil {
							debug.Erro("etcd create key[%s] value[%s] failure, error: %s", string(ev.Kv.Key), string(ev.Kv.Value), err)
						}
					} else if ev.IsModify() {
						if err := e.Modify(string(ev.Kv.Key), string(ev.Kv.Value)); err != nil {
							debug.Erro("etcd modify key[%s] value[%s] failure, error: %s", string(ev.Kv.Key), string(ev.Kv.Value), err)
						}
					}
				case mvccpb.DELETE:
					if err := e.Delete(string(ev.Kv.Key), string(ev.Kv.Value)); err != nil {
						debug.Erro("etcd delete key[%s] value[%s] failure, error: %s", string(ev.Kv.Key), string(ev.Kv.Value), err)
					}
				}
			}
		}
	}
}

func (c *Client) Shutdown() {
	c.locker.RLock()
	defer c.locker.RUnlock()

	for _, cancel := range c.cancels {
		cancel()
	}

	if c.cli != nil {
		if err := c.cli.Close(); err != nil {
			debug.Erro("etcd close failure, error: %s", err)
		}
	}

	c.wait.Wait()
}

func (c *Client) Close(key string) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if cancel, ok := c.cancels[key]; ok {
		cancel()
		delete(c.cancels, key)
	}
}

func (c *Client) Revoke(leaseId clientv3.LeaseID) error {
	ctx, cancel := context.WithTimeout(context.Background(), Req_Timeout*time.Second)
	defer cancel()
	_, err := c.cli.Revoke(ctx, leaseId)
	return err
}

func (c *Client) Grant(ttl int64) (clientv3.LeaseID, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Req_Timeout*time.Second)
	defer cancel()
	res, err := c.cli.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}

	return res.ID, nil
}

func (c *Client) KeepAlive(leaseId clientv3.LeaseID) (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Req_Timeout*time.Second)
	defer cancel()

	return c.cli.KeepAlive(ctx, leaseId)
}

func (c *Client) KeepAliveOnce(leaseId clientv3.LeaseID) (*clientv3.LeaseKeepAliveResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), Req_Timeout*time.Second)
	defer cancel()

	return c.cli.KeepAliveOnce(ctx, leaseId)
}
