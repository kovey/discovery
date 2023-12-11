package krpc

import (
	"testing"

	"github.com/kovey/discovery/algorithm"
)

func TestLoadBalance(t *testing.T) {
	load := NewLoadBalance(algorithm.Alg_Random)
	if load.encode() != `{"loadBalancingConfig":[{"random":{}}]}` {
		t.Fatalf("test failure")
	}
}
