package register

import (
	"net"
	"strings"

	"github.com/kovey/debug-go/debug"
)

func InnerIp() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		debug.Erro("get addrs failure, error: %s", err)
		return ""
	}

	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}

	return ""
}

func OutIp() string {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		debug.Erro("get addrs failure, error: %s", err)
		return ""
	}

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return strings.Split(localAddr.String(), ":")[0]
}
