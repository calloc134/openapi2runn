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
	"github.com/k0kubun/pp"
)

func generateDirName(str, sep string) string {

	re := regexp.MustCompile("[@{}]+")
	str_noSP := re.ReplaceAllString(str, "")
	//fmt.Println(str_noSP)
	parts := strings.Split(str_noSP, sep)
	for i, part := range parts {
		parts[i] = cases.Title(language.Und, cases.NoLower).String(part)
	}
	return strings.Join(parts, "")
}

func main() {

	type paramSpec struct {
		Name string `json:"name"`
		Type string `json:"type"`
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

	var pathSpecs []pathSpec

	doc, err := openapi3.NewLoader().LoadFromFile("openapi.yaml")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	for _, path := range doc.Paths.InMatchingOrder() {

		//fmt.Println(path)

		obj := doc.Paths.Find(path).Operations()

		var baseSpecs []baseSpec

		for i, op := range obj {

			var queries []paramSpec
			var bodies []paramSpec

			if op.Parameters != nil {
				for _, queryParam := range op.Parameters {

					queries = append(queries, paramSpec{
						Name: queryParam.Value.Name,
						Type: queryParam.Value.Schema.Value.Type,
					})
				}
			}

			if op.RequestBody != nil {
				for name, bodyParams := range op.RequestBody.Value.Content["application/json"].Schema.Value.Properties {
					bodies = append(bodies, paramSpec{
						Name: name,
						Type: bodyParams.Value.Type,
					})
				}
			}

			baseSpecs = append(baseSpecs, baseSpec{
				Method: i,
				Body:   bodies,
				Params: queries,
			})

		}

		pathSpecs = append(pathSpecs, pathSpec{
			DirName: generateDirName(path, "/"),
			Path:    path,
			Methods: baseSpecs,
		})

	}

	//pp.Print(pathSpecs)

	//yamlData, err := yaml.Marshal(pathSpecs)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// YAML形式のデータを表示する
	//fmt.Println(string(yamlData))

	for _, pathSpec := range pathSpecs {

		for _, method := range pathSpec.Methods {
			if err := os.MkdirAll("swagger/0_base/"+pathSpec.DirName+"/"+method.Method, 0777); err != nil {
				fmt.Println(err)
			}

			if err := os.MkdirAll("swagger/1_noAuth/"+pathSpec.DirName+"/"+method.Method, 0777); err != nil {
				fmt.Println(err)
			}

			tmpl_base, err := template.New("base.yml.template").Delims("<<", ">>").ParseFiles("template/base.yml.template")

			if err != nil {
				fmt.Println(err)
			}

			re := regexp.MustCompile(`{\s*([^}\s]*)\s*}`)

			alterPath := re.ReplaceAllString(pathSpec.Path, `{{ vars.req.query.$1 }}`)

			fp_base, err := os.Create("swagger/0_base/" + pathSpec.DirName + "/" + method.Method + "/base.yml")
			if err != nil {
				fmt.Println(err)
			}
			defer fp_base.Close()

			outputBodies := ""

			if method.Body != nil {
				outputBodies = "application/json: \"{{ vars.req.body }}\""
			} else if method.Method != "get" {
				outputBodies = "application/json: []"
			} else {
				outputBodies = "null"
			}

			err = tmpl_base.Execute(fp_base, map[string]any{
				"method": method.Method,
				"path":   alterPath,
				"bodies": outputBodies,
			})

			if err != nil {
				fmt.Println(err)
			}

			tmpl_noAuth, err := template.New("index.yml.template").Delims("<<", ">>").ParseFiles("template/index.yml.template")

			fp_noAuth, err := os.Create("swagger/1_noAuth/" + pathSpec.DirName + "/" + method.Method + "/base.yml")
			if err != nil {
				fmt.Println(err)
			}
			defer fp_noAuth.Close()

			err = tmpl_noAuth.Execute(fp_noAuth, map[string]any{
				"desc":    "(" + method.Method + ") " + pathSpec.Path + "のテスト",
				"method":  method.Method,
				"dirname": pathSpec.DirName,
			})

			if err != nil {
				fmt.Println(err)
			}

			tmpl_data, err := template.New("data.json.template").Delims("<<", ">>").ParseFiles("template/data.json.template")

			fp_data, err := os.Create("swagger/1_noAuth/" + pathSpec.DirName + "/" + method.Method + "/data.json")

			if err != nil {
				fmt.Println(err)
			}

			jsonBodyMap := map[string]any{}

			for _, body := range method.Body {

				if body.Type == "string" {
					jsonBodyMap[body.Name] = "dummy"
				} else if body.Type == "number" {
					jsonBodyMap[body.Name] = 0
				} else {
					jsonBodyMap[body.Name] = ""
				}
			}
			jsonBodyObj, err := json.Marshal(jsonBodyMap)

			jsonBody := string(jsonBodyObj)

			jsonQueryMap := map[string]any{}

			for _, query := range method.Params {

				if query.Type == "string" {
					jsonQueryMap[query.Name] = "dummy"
				} else if query.Type == "number" {
					jsonQueryMap[query.Name] = 0
				} else {
					jsonQueryMap[query.Name] = ""
				}

			}

			pp.Println(jsonQueryMap)

			jsonQueryObj, err := json.Marshal(jsonQueryMap)
			jsonQuery := string(jsonQueryObj)

			defer fp_data.Close()
			err = tmpl_data.Execute(fp_data, map[string]any{
				"jsonBody":  jsonBody,
				"jsonQuery": jsonQuery,
			})

		}
	}

}
