package algorithm

import (
	"context"
	"math/rand"
	"testing"

	"github.com/kovey/discovery/grpc"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

func TestRandomWeightPicker(t *testing.T) {
	picker := &randomWeightPicker{subConns: make([]balancer.SubConn, 2), rand: rand.New(rand.NewSource(100)), instances: make([]*weightInfo, 2), total: 100}
	picker.subConns[0] = &testSubConn{name: "35%"}
	picker.subConns[1] = &testSubConn{name: "65%"}
	picker.instances[0] = &weightInfo{min: 0, max: 34, instance: &grpc.Instance{Name: "test.test", Addr: "127.0.0.1:9901", Version: "1.0", Weight: 35}}
	picker.instances[1] = &weightInfo{min: 35, max: 99, instance: &grpc.Instance{Name: "test.test", Addr: "127.0.0.1:9902", Version: "1.0", Weight: 65}}
	res, err := picker.Pick(balancer.PickInfo{Ctx: context.Background(), FullMethodName: "test.Test.Test"})
	if err != nil {
		t.Fatalf("test picker failure, error: %s", err)
	}

	if res.SubConn == nil {
		t.Fatalf("test picker failure, sub conn is nil")
	}
	sub, ok := res.SubConn.(*testSubConn)
	if !ok {
		t.Fatalf("conn not *testSubConn")
	}

	if sub.name != "65%" {
		t.Fatal("conn is not 65%")
	}
}

func TestRandomWeight(t *testing.T) {
	builder := &randomWeightBuilder{}
	subConns := make(map[balancer.SubConn]base.SubConnInfo)
	ins := &grpc.Instance{Name: "test.Test", Addr: "127.0.0.1:9901", Version: "1.0", Weight: 35}
	subConns[&testSubConn{name: "35%"}] = base.SubConnInfo{Address: ins.Address()}
	ins = &grpc.Instance{Name: "test.Test", Addr: "127.0.0.1:9902", Version: "1.0", Weight: 65}
	subConns[&testSubConn{name: "65%"}] = base.SubConnInfo{Address: ins.Address()}
	picker := builder.Build(base.PickerBuildInfo{ReadySCs: subConns})
	res, err := picker.Pick(balancer.PickInfo{Ctx: context.Background(), FullMethodName: "test.Test"})
	if err != nil {
		t.Fatalf("test random builder failure, error: %s", err)
	}

	sub, ok := res.SubConn.(*testSubConn)
	if !ok {
		t.Fatalf("test random builder failure, conn not *testSubConn")
	}

	t.Logf("conn is %s", sub.name)
}
