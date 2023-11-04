package krpc

import (
	"fmt"

	dg "github.com/kovey/discovery/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var mode Mode = Mode_Remote

func SetMode(m Mode) {
	mode = m
}

func dial(serviceName string) (*grpc.ClientConn, error) {
	switch mode {
	case Mode_Local:
		local, ok := locals.Get(serviceName)
		if !ok {
			return nil, fmt.Errorf("service[%s] not found on local", serviceName)
		}
		return grpc.Dial(fmt.Sprintf("%s:%d", local.Host, local.Port), grpc.WithTransportCredentials(insecure.NewCredentials()))
	default:
		return grpc.Dial(
			fmt.Sprintf("%s:///%s", dg.Scheme_Etcd, serviceName), grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		)
	}
}

func Dial(serviceName ServiceName) (grpc.ClientConnInterface, error) {
	if conn, err := c.get(serviceName.Default()); err == nil {
		return conn, err
	}

	return c.add(serviceName.Default())
}

func DialWithGroup(serviceName ServiceName, group string) (grpc.ClientConnInterface, error) {
	if conn, err := c.get(serviceName.Group(group)); err == nil {
		return conn, err
	}

	return c.add(serviceName.Group(group))
}

func Close() {
	c.close()
}
