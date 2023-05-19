package plugin

import (
	"encoding/json"
	"fmt"
	"strings"
)

var ErrNotCmd = fmt.Errorf("not found cmd")

const CMDVALPREFIX = "CMDVAL:"
const CMDERRORREFIX = "CMDERROR:"
const CMDSTATEPREFIX = "CMDSTATE:"

const COMPLETELOG = "COMPLETED"

func respError(err error) {
	fmt.Print(CMDERRORREFIX)
	fmt.Println(err.Error())
}

func respJson(val interface{}) {
	data, err := json.Marshal(val)
	if err != nil {
		respError(err)
	}
	fmt.Print(CMDVALPREFIX)
	fmt.Println(string(data))
}

func respState(state string) {
	fmt.Print(CMDSTATEPREFIX)
	fmt.Println(state)
}

func isCmd(cmd string) bool {
	return cmd == CMDVALPREFIX || cmd == CMDERRORREFIX || cmd == CMDSTATEPREFIX
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
