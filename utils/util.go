package utils

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func Prompt() {
	fmt.Printf("-> Press Return key to continue.")
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		break
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	fmt.Println()
}

func Int32Ptr(i int32) *int32 { return &i }

func StringPtr(str string) *string { return &str }

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close() //nolint
	return l.Addr().(*net.TCPAddr).Port, nil
}

func ToMultiAddr(endpoint string) string {
	seq := strings.Split(endpoint, ":")
	port := seq[1]
	ipOrDsn := seq[0]
	addr := net.ParseIP(ipOrDsn)
	if addr != nil {
		return fmt.Sprintf("/ip4/%s/tcp/%s", ipOrDsn, port)
	}

	return fmt.Sprintf("/dsn/%s/tcp/%s", ipOrDsn, port)
}

func HasDupItemInArrary(arr []string) bool {
	filter := make(map[string]bool)
	for _, v := range arr {
		_, ok := filter[v]
		if !ok {
			filter[v] = true
			continue
		}
		return true
	}
	return false
}

func DistinctArrary(arr []string) []string {
	var result []string
	filter := make(map[string]struct{})
	for _, v := range arr {
		_, ok := filter[v]
		if !ok {
			filter[v] = struct{}{}
			result = append(result, v)
			continue
		}
	}
	return result
}
