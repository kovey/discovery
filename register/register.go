package register

import (
	"fmt"

	"github.com/kovey/discovery/etcd"
	"github.com/kovey/discovery/grpc"
)

var reg *grpc.Register

func Init(conf etcd.Config) {
	if reg != nil {
		return
	}

	reg = grpc.NewRegister(conf)
}

func Register(ins *grpc.Instance, ttl int64) error {
	if reg == nil {
		return fmt.Errorf("register not init")
	}

	return reg.Register(ins, ttl)
}

func Shutdown() {
	if reg == nil {
		return
	}

	reg.Shutdown()
}

func GetInstance() (*grpc.Instance, error) {
	if reg == nil {
		return nil, fmt.Errorf("register not init")
	}

	return reg.GetInstance()
}
