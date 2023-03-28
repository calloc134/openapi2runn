package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"os"

	"github.com/getkin/kin-openapi/openapi3"
)

type paramSpec struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Example any
}

type baseSpec struct {
	Method string
	Params []paramSpec
	Body   []paramSpec
}

type pathSpec struct {
	DirName string
	Path    string
	Methods []baseSpec
}

func genDirName(str, sep string) string {

	re := regexp.MustCompile("[@{}]+")
	str_noSP := re.ReplaceAllString(str, "")
	//fmt.Println(str_noSP)
	parts := strings.Split(str_noSP, sep)
	for i, part := range parts {
		parts[i] = cases.Title(language.Und, cases.NoLower).String(part)
	}
	return strings.Join(parts, "")
}

func genJson(paramSpecs []paramSpec) string {

	jsonBodyMap := map[string]any{}
	for _, param := range paramSpecs {
		if param.Type == "string" {
			if param.Example != nil {
				jsonBodyMap[param.Name] = param.Example
			} else {
				jsonBodyMap[param.Name] = "dummy"
			}
		} else if param.Type == "number" {
			jsonBodyMap[param.Name] = 0
		} else {
			jsonBodyMap[param.Name] = ""
		}
	}
	jsonBodyObj, err := json.Marshal(jsonBodyMap)
	if err != nil {
		fmt.Println(err)
	}

	return string(jsonBodyObj)

}

func main() {

	var pathSpecs []pathSpec

	doc, err := openapi3.NewLoader().LoadFromFile("openapi.yaml")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, path := range doc.Paths.InMatchingOrder() {

		obj := doc.Paths.Find(path).Operations()

		var baseSpecs []baseSpec

		for method, op := range obj {

			var queries []paramSpec
			var bodies []paramSpec

			if op.Parameters != nil {
				for _, q := range op.Parameters {

					queries = append(queries, paramSpec{
						Name:    q.Value.Name,
						Type:    q.Value.Schema.Value.Type,
						Example: q.Value.Example,
					})
				}
			}

			if op.RequestBody != nil {

				for name, b := range op.RequestBody.Value.Content["application/json"].Schema.Value.Properties {
					bodies = append(bodies, paramSpec{
						Name:    name,
						Type:    b.Value.Type,
						Example: b.Value.Example,
					})
				}
			}

			baseSpecs = append(baseSpecs, baseSpec{
				Method: method,
				Body:   bodies,
				Params: queries,
			})

		}

		pathSpecs = append(pathSpecs, pathSpec{
			DirName: genDirName(path, "/"),
			Path:    path,
			Methods: baseSpecs,
		})

	}

	for _, pathSpec := range pathSpecs {

		for _, methodItem := range pathSpec.Methods {

			re := regexp.MustCompile(`{\s*([^}\s]*)\s*}`)
			modiPath := re.ReplaceAllString(pathSpec.Path, `{{ vars.req.query.$1 }}`)

			if err := os.MkdirAll("swagger/0_base/"+pathSpec.DirName+"/"+methodItem.Method, 0777); err != nil {
				fmt.Println(err)
			}
			if err := os.MkdirAll("swagger/1_noAuth/"+pathSpec.DirName+"/"+methodItem.Method, 0777); err != nil {
				fmt.Println(err)
			}
			tmpl_base, err := template.New("base.yml.template").Delims("<<", ">>").ParseFiles("template/base.yml.template")
			if err != nil {
				fmt.Println(err)
			}

			fp_base, err := os.Create("swagger/0_base/" + pathSpec.DirName + "/" + methodItem.Method + "/base.yml")
			if err != nil {
				fmt.Println(err)
			}
			defer fp_base.Close()

			modiBody := ""
			if methodItem.Body != nil {
				modiBody = "application/json: \"{{ vars.req.body }}\""
			} else if methodItem.Method != "get" {
				modiBody = "application/json: []"
			} else {
				modiBody = "null"
			}

			err = tmpl_base.Execute(fp_base, map[string]any{
				"method": methodItem.Method,
				"path":   modiPath,
				"bodies": modiBody,
			})
			if err != nil {
				fmt.Println(err)
			}

			tmpl_noAuth, err := template.New("index.yml.template").Delims("<<", ">>").ParseFiles("template/index.yml.template")
			if err != nil {
				fmt.Println(err)
			}

			fp_noAuth, err := os.Create("swagger/1_noAuth/" + pathSpec.DirName + "/" + methodItem.Method + "/base.yml")
			if err != nil {
				fmt.Println(err)
			}
			defer fp_noAuth.Close()

			err = tmpl_noAuth.Execute(fp_noAuth, map[string]any{
				"desc":    "(" + methodItem.Method + ") " + pathSpec.Path + "のテスト",
				"method":  methodItem.Method,
				"dirname": pathSpec.DirName,
			})

			if err != nil {
				fmt.Println(err)
			}

			tmpl_data, err := template.New("data.json.template").Delims("<<", ">>").ParseFiles("template/data.json.template")
			if err != nil {
				fmt.Println(err)
			}

			fp_data, err := os.Create("swagger/1_noAuth/" + pathSpec.DirName + "/" + methodItem.Method + "/data.json")
			if err != nil {
				fmt.Println(err)
			}
			defer fp_data.Close()

			jsonBody := genJson(methodItem.Body)
			jsonQuery := genJson(methodItem.Params)

			if err = tmpl_data.Execute(fp_data, map[string]any{
				"jsonBody":  jsonBody,
				"jsonQuery": jsonQuery,
			}); err != nil {
				fmt.Println(err)
			}

		}
	}

}
