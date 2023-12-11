package krpc

import (
	"fmt"

	"github.com/kovey/discovery/grpc"
	"github.com/kovey/discovery/register"
)

type ServiceName string

func (s ServiceName) Group(group string) string {
	if group == grpc.Str_Empty {
		group = grpc.Default
	}

	return fmt.Sprintf(service_name, grpc.Namespace(), group, s)
}

func (s ServiceName) Default() string {
	return s.Group(grpc.Default)
}

type Local struct {
	Host    string      `yaml:"host" json:"host"`
	Port    int         `yaml:"port" json:"port"`
	Name    ServiceName `yaml:"name" json:"name"`
	Group   string      `yaml:"group" json:"group"`
	Weight  int64       `yaml:"weight" json:"weight"`
	Version string      `yaml:"version" json:"version"`
}

func (l *Local) Addr() string {
	return fmt.Sprintf("%s:%d", l.Host, l.Port)
}

func (l *Local) ServiceName() string {
	if l.Group == grpc.Str_Empty {
		l.Group = grpc.Default
	}
	return l.Name.Group(l.Group)
}

func (l *Local) InnerAddr() string {
	return fmt.Sprintf("%s:%d", register.InnerIp(), l.Port)
}

func (l *Local) Instance() *grpc.Instance {
	if l.Group == grpc.Str_Empty {
		l.Group = grpc.Default
	}

	if l.Weight == 0 {
		l.Weight = 100
	}

	return &grpc.Instance{Name: string(l.Name), Addr: l.InnerAddr(), Version: l.Version, Group: l.Group, Namespace: grpc.Namespace(), Weight: l.Weight}
}

type Locals map[string]*Local

func (l Locals) Get(name string) (*Local, bool) {
	local, ok := l[name]
	return local, ok
}

func (l Locals) Add(lo *Local) {
	l[lo.ServiceName()] = lo
}

var locals = make(Locals)

func AddLocal(local *Local) {
	locals.Add(local)
}

func SetLocals(locals []*Local) {
	for _, local := range locals {
		AddLocal(local)
	}
}
