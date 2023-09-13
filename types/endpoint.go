package types

import (
	"fmt"
	"net"
	"strconv"
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
	if len(endpoint) == 0 {
		return ""
	}

	seq := strings.Split(string(endpoint), ":")
	port := seq[1]
	ipOrDNS := seq[0]
	addr := net.ParseIP(ipOrDNS)
	if addr != nil {
		return fmt.Sprintf("/ip4/%s/tcp/%s", ipOrDNS, port)
	}
	return fmt.Sprintf("/dns/%s/tcp/%s", ipOrDNS, port)
}

func (endpoint Endpoint) ToHTTP() string {
	if len(endpoint) == 0 {
		return ""
	}
	return fmt.Sprintf("http://%s", string(endpoint))
}

func (endpoint Endpoint) IP() string {
	return strings.Split(string(endpoint), ":")[0]
}

func (endpoint Endpoint) Port() int {
	port, err := strconv.Atoi(strings.Split(string(endpoint), ":")[1])
	if err != nil {
		panic("invalida endpoint " + endpoint)
	}
	return port
}
