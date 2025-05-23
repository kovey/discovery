package krpc

import (
	"fmt"

	"github.com/kovey/discovery/algorithm"
	dg "github.com/kovey/discovery/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var mode = Mode_Remote
var loadBalance = NewLoadBalance(algorithm.Alg_Round_Robin)

func SetMode(m Mode) {
	mode = m
}

func SetLoadBalance(name string) {
	loadBalance = NewLoadBalance(name)
}

func dial(serviceName string) (*grpc.ClientConn, error) {
	switch mode {
	case Mode_Local:
		local, ok := locals.Get(serviceName)
		if !ok {
			return nil, fmt.Errorf("service[%s] not found on local", serviceName)
		}
		return grpc.Dial(fmt.Sprintf("%s:%d", local.Host, local.Port), grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithChainUnaryInterceptor(trace))
	default:
		return grpc.Dial(
			fmt.Sprintf("%s://%s", dg.Scheme_Etcd, serviceName), grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDefaultServiceConfig(loadBalance.encode()), grpc.WithChainUnaryInterceptor(trace),
		)
	}
}

func Dial(serviceName ServiceName, group string) (grpc.ClientConnInterface, error) {
	if conn, err := c.get(serviceName.Group(group)); err == nil {
		return conn, err
	}

	return c.add(serviceName.Group(group))
}

func DialDefault(serviceName ServiceName) (grpc.ClientConnInterface, error) {
	return Dial(serviceName, dg.Default)
}

func Close() {
	c.close()
}
