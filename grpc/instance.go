package grpc

import (
	"encoding/json"
	"fmt"
	"strings"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

const (
	ins_prefix         = "/%s/%s/%s/"
	ins_prefix_version = "/%s/%s/%s/%s/"
	ins_path           = "%s%s"
	ins_index          = "/"
	Str_Empty          = ""
	addr_split         = ":"
	ins_weight         = "weight"
	Default            = "default"
	key_group          = "group"
	key_version        = "version"
	key_namespace      = "namespace"
)

type Instance struct {
	Name      string `json:"n"`
	Addr      string `json:"a"`
	Version   string `json:"v"`
	Weight    int64  `json:"w"`
	Group     string `json:"g"`
	Namespace string `json:"ns"`
}

func (i *Instance) Valid() bool {
	return strings.Split(i.Addr, addr_split)[0] != Str_Empty
}

func (i *Instance) Prefix() string {
	if i.Group == Str_Empty {
		i.Group = Default
	}
	if i.Namespace == Str_Empty {
		i.Namespace = Default
	}

	if i.Version == Str_Empty {
		return fmt.Sprintf(ins_prefix, i.Namespace, i.Group, i.Name)
	}

	return fmt.Sprintf(ins_prefix_version, i.Namespace, i.Group, i.Name, i.Version)
}

func (i *Instance) Path() string {
	return fmt.Sprintf(ins_path, i.Prefix(), i.Addr)
}

func (i *Instance) Parse(addr resolver.Address) {
	i.Addr = addr.Addr
	i.Name = addr.ServerName
	i.Version = addr.Attributes.Value(key_version).(string)
	i.Weight = addr.Attributes.Value(ins_weight).(int64)
	i.Group = addr.Attributes.Value(key_group).(string)
	i.Namespace = addr.Attributes.Value(key_namespace).(string)
}

func (i *Instance) Decode(value string) error {
	return json.Unmarshal([]byte(value), i)
}

func (i *Instance) Encode() (string, error) {
	buf, err := json.Marshal(i)
	if err == nil {
		return string(buf), nil
	}

	return Str_Empty, err
}

func (i *Instance) Address() resolver.Address {
	return resolver.Address{
		Addr: i.Addr, ServerName: i.Name,
		Attributes: attributes.New(ins_weight, i.Weight).WithValue(key_group, i.Group).WithValue(key_version, i.Version).WithValue(key_namespace, i.Namespace),
	}
}
