package plugin

import (
	"fmt"
	"strings"
)

var ErrNotCmd = fmt.Errorf("not found cmd")

const CMDERRORREFIX = "CMDERROR:"
const CMDSTARTPREFIX = "CMDSTART:"
const CMDSUCCESSPREFIX = "CMDSUCCESS:"

func RespError(err error) {
	fmt.Print(CMDERRORREFIX)
	fmt.Println(err.Error())
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
	return cmd == CMDERRORREFIX || cmd == CMDSTARTPREFIX || cmd == CMDSUCCESSPREFIX
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

func GetLastJSON(text string) string {
	text = strings.Trim(text, "\n \t")
	if text[len(text)-1] != '}' {
		return ""
	}

	isClose := 0
	for i := len(text) - 1; i >= 0; i-- {
		switch text[i] {
		case '}':
			isClose++
		case '{':
			isClose--
		}
		if isClose == 0 {
			return text[i:]
		}
	}
	return ""
}
