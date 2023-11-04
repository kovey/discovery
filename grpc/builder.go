package grpc

import (
	"github.com/kovey/discovery/etcd"
	"google.golang.org/grpc/resolver"
)

const (
	Scheme_Etcd = "etcd"
)

type Builder struct {
	cli    *etcd.Client
	scheme string
	conf   etcd.Config
}

func NewBuilder(conf etcd.Config) *Builder {
	return &Builder{cli: etcd.NewClient(), scheme: Scheme_Etcd, conf: conf}
}

func (b *Builder) Register() error {
	if err := b.cli.Connect(b.conf); err != nil {
		return err
	}

	resolver.Register(b)
	return nil
}

func (b *Builder) Shutdown() {
	b.cli.Shutdown()
}

func (b *Builder) Scheme() string {
	return b.scheme
}

func (b *Builder) Build(target resolver.Target, conn resolver.ClientConn, opt resolver.BuildOptions) (resolver.Resolver, error) {
	r := NewResolver(conn, b.cli, target.URL.Path, true)
	if err := r.start(); err != nil {
		return nil, err
	}

	return r, nil
}
