package krpc

type Mode string

const (
	Mode_Local  Mode = "local"  // local service
	Mode_Remote Mode = "remote" // remote service
)

const (
	Scheme_Etcd  = "etcd"
	service_name = "%s/%s"
)
