package templating

import (
	"bytes"
	"encoding/json"
	"io"

	"os"
	"text/template"
)

func add(a, b int) int                          { return a + b }
func sub(a, b int) int                          { return a - b }
func hasNext(array []interface{}, idx int) bool { return idx < len(array)-1 }
func isLast(array []interface{}, idx int) bool  { return idx == len(array)-1 }

func ParseTemplateJsonData(templ string, data string) (string, error) {

	m := map[string]interface{}{}

	if data != "" {
		// parse data json file to map

		jsonFile, err := os.Open(data)
		if err != nil {
			return "", err
		}

		dataBytes, _ := io.ReadAll(jsonFile)
		if err := json.Unmarshal(dataBytes, &m); err != nil {
			return "", err
		}
	} else {
		// parse empty json as data
		if err := json.Unmarshal([]byte("{ }"), &m); err != nil {
			return "", err
		}
	}

	templateFile, _ := os.Open(templ)
	templateBytes, _ := io.ReadAll(templateFile)

	funcMap := template.FuncMap{
		"add":     add,
		"sub":     sub,
		"hasNext": hasNext,
		"isLast":  isLast,
	}

	t, err := template.New("").Funcs(funcMap).Parse(string(templateBytes))
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, m); err != nil {
		return "", err
	}

	return tpl.String(), nil
}
