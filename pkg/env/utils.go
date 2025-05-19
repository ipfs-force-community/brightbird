package env

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strings"
	"text/template"

	v1 "k8s.io/api/core/v1"
)

func GetPodDNS(svc *v1.Service, pods ...v1.Pod) []string {
	podDNS := make([]string, len(pods))
	for index, pod := range pods {
		podDNS[index] = fmt.Sprintf("%s.%s", pod.GetName(), svc.GetName())
	}
	return podDNS
}

func QuickRender(file fs.File, args any) ([]byte, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	render, err := template.New("").Funcs(template.FuncMap{
		"join": func(s []string, sep string) string {
			return strings.Join(s, sep)
		},
	}).Parse(string(data))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer(nil)
	err = render.Execute(buf, args)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
