package plugin

import (
	"encoding/json"
	"fmt"
	"strings"
)

var ErrNotCmd = fmt.Errorf("not found cmd")

const CMDVALPREFIX = "CMDVAL:"
const CMDERRORREFIX = "CMDERROR:"
const CMDSTARTPREFIX = "CMDSTART:"
const CMDSUCCESSPREFIX = "CMDSUCCESS:"

func RespError(err error) {
	fmt.Print(CMDERRORREFIX)
	fmt.Println(err.Error())
}

func RespJSON(val interface{}) {
	data, err := json.Marshal(val)
	if err != nil {
		RespError(err)
	}
	fmt.Print(CMDVALPREFIX)
	fmt.Println(string(data))
}

func RespStart(addition string) {
	fmt.Print(CMDSTARTPREFIX)
	fmt.Println(addition)
}

func RespSuccess(addition string) {
	fmt.Print(CMDSUCCESSPREFIX)
	fmt.Println(addition)
}

func isCmd(cmd string) bool {
	return cmd == CMDVALPREFIX || cmd == CMDERRORREFIX || cmd == CMDSTARTPREFIX || cmd == CMDSUCCESSPREFIX
}

func ReadCMD(line string) (string, string, bool) {
	line = strings.Trim(line, "\n")
	cmd := ""
	val := ""
	for pos, char := range line {
		if char == ':' {
			cmd = line[:pos+1]
			val = line[pos+1:]
			break
		}
	}
	if !isCmd(cmd) {
		return "", "", false
	}
	return cmd, val, true
}
