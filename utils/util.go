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
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func ToMultiAddr(endpoint string) string {
	seq := strings.Split(endpoint, ":")
	port := seq[1]
	ipOrDsn := seq[0]
	addr := net.ParseIP(ipOrDsn)
	if addr != nil {
		return fmt.Sprintf("/ip4/%s/tcp/%s", ipOrDsn, port)
	} else {
		return fmt.Sprintf("/dsn/%s/tcp/%s", ipOrDsn, port)
	}
}
