package krpc

import (
	"fmt"

	"github.com/kovey/discovery/grpc"
)

type ServiceName string

func (s ServiceName) Group(namespace, group string) string {
	return fmt.Sprintf(service_name, namespace, group, s)
}

func (s ServiceName) Default() string {
	return s.Group(grpc.Default, grpc.Default)
}

type Local struct {
	Host      string      `yaml:"host" json:"host"`
	Port      int         `yaml:"port" json:"port"`
	Name      ServiceName `yaml:"name" json:"name"`
	Group     string      `yaml:"group" json:"group"`
	Namespace string      `yaml:"namespace" json:"namespace"`
	Weight    int64       `yaml:"weight" json:"weight"`
	Version   string      `yaml:"version" json:"version"`
}

func (l *Local) Addr() string {
	return fmt.Sprintf("%s:%d", l.Host, l.Port)
}

func (l *Local) ServiceName() string {
	if l.Group == grpc.Str_Empty {
		l.Group = grpc.Default
	}
	if l.Namespace == grpc.Str_Empty {
		l.Namespace = grpc.Default
	}

	return l.Name.Group(l.Namespace, l.Group)
}

func (l *Local) Instance() *grpc.Instance {
	return &grpc.Instance{Name: string(l.Name), Addr: l.Addr(), Version: l.Version, Group: l.Group, Namespace: l.Namespace}
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
