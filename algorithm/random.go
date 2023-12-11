package algorithm

import (
	"math/rand"
	"time"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
)

const (
	Alg_Round_Robin = "round_robin"
	Alg_Random      = "random"
)

var logger = grpclog.Component(Alg_Random)

func init() {
	balancer.Register(newRandomBuilder())
}

type randomBuilder struct {
}

func newRandomBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Alg_Random, &randomBuilder{}, base.Config{HealthCheck: true})
}

func (r *randomBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		logger.Errorf("sub connection is empty")
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	subConns := make([]balancer.SubConn, len(info.ReadySCs))
	index := 0
	for subConn := range info.ReadySCs {
		subConns[index] = subConn
		index++
	}

	return &randomPicker{subConns: subConns, rand: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

type randomPicker struct {
	subConns []balancer.SubConn
	rand     *rand.Rand
}

func (r *randomPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if err := info.Ctx.Err(); err != nil {
		logger.Errorf("context is error: %s", err)
		return balancer.PickResult{}, err
	}

	next := r.rand.Intn(len(r.subConns))
	return balancer.PickResult{SubConn: r.subConns[next]}, nil
}
