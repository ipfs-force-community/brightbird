package types

import (
	"fmt"
	"net"
	"strings"
)

type Endpoint string

func EndpointFromString(str string) Endpoint {
	return Endpoint(str)
}

func EndpointFromHostPort(host string, port int) Endpoint {
	return Endpoint(fmt.Sprintf("%s:%d", host, port))
}

func (endpoint Endpoint) ToMultiAddr() string {
	seq := strings.Split(string(endpoint), ":")
	port := seq[1]
	ipOrDns := seq[0]
	addr := net.ParseIP(ipOrDns)
	if addr != nil {
		return fmt.Sprintf("/ip4/%s/tcp/%s", ipOrDns, port)
	} else {
		return fmt.Sprintf("/dns/%s/tcp/%s", ipOrDns, port)
	}
}

func (endpoint Endpoint) ToHttp() string {
	return fmt.Sprintf("http://%s", string(endpoint))
}
