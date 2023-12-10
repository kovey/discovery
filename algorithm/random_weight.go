package algorithm

import (
	"math/rand"
	"time"

	"github.com/kovey/discovery/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/grpclog"
)

const (
	Alg_Random_Weight = "random_weight"
)

var log = grpclog.Component(Alg_Random_Weight)

func init() {
	balancer.Register(newRandomWeightBuilder())
}

type randomWeightBuilder struct {
}

func newRandomWeightBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Alg_Random_Weight, &randomWeightBuilder{}, base.Config{HealthCheck: true})
}

func (r *randomWeightBuilder) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		log.Errorf("sub connection is empty")
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	subConns := make([]balancer.SubConn, len(info.ReadySCs))
	instances := make([]*weightInfo, len(info.ReadySCs))
	index := 0
	total := int64(0)
	for subConn, addr := range info.ReadySCs {
		subConns[index] = subConn
		instance := &grpc.Instance{}
		instance.Parse(addr.Address)
		wInfo := &weightInfo{instance: instance, min: total}
		total += instance.Weight
		wInfo.max = total - 1
		instances[index] = wInfo
		index++
	}

	if total <= 0 {
		return base.NewErrPicker(balancer.ErrBadResolverState)
	}

	return &randomWeightPicker{subConns: subConns, rand: rand.New(rand.NewSource(time.Now().UnixNano())), instances: instances, total: total}
}

type weightInfo struct {
	instance *grpc.Instance
	min      int64
	max      int64
}

type randomWeightPicker struct {
	subConns  []balancer.SubConn
	rand      *rand.Rand
	instances []*weightInfo
	total     int64
}

func (r *randomWeightPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	if err := info.Ctx.Err(); err != nil {
		log.Errorf("context is error: %s", err)
		return balancer.PickResult{}, err
	}

	next := r.rand.Int63n(r.total)
	for index, wInfo := range r.instances {
		if next >= wInfo.min && next <= wInfo.max {
			return balancer.PickResult{SubConn: r.subConns[index]}, nil
		}
	}

	log.Errorf("random weight failure, total[%d], random[%d]", r.total, next)
	return balancer.PickResult{}, balancer.ErrNoSubConnAvailable
}
