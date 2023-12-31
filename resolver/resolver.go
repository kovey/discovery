package resolver

import (
	"fmt"

	_ "github.com/kovey/discovery/algorithm"
	"github.com/kovey/discovery/etcd"
	"github.com/kovey/discovery/grpc"
)

var builder *grpc.Builder

func Init(conf etcd.Config) {
	grpc.SetNamespace(conf.Namespace)
	builder = grpc.NewBuilder(conf)
}

func Register() error {
	if builder == nil {
		return fmt.Errorf("builder not init")
	}

	return builder.Register()
}

func Shutdown() {
	if builder == nil {
		return
	}

	builder.Shutdown()
}
