package etcd

import "time"

type Config struct {
	Endpoints   []string `yaml:"endpoints" json:"endpoints"`
	DialTimeout int      `yaml:"timeout" json:"timeout"`
	Username    string   `yaml:"username" json:"username"`
	Password    string   `yaml:"password" json:"password"`
}

const (
	Req_Timeout time.Duration = 30
)
