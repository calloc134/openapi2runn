package cmd

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"os"

	"github.com/BurntSushi/toml"
	"github.com/getkin/kin-openapi/openapi3"
)

// ディレクトリ名を生成する関数
func genDirName(str, sep string) string {

	re := regexp.MustCompile("[@{}]+")
	str_noSP := re.ReplaceAllString(str, "")
	parts := strings.Split(str_noSP, sep)
	for i, part := range parts {
		parts[i] = cases.Title(language.Und, cases.NoLower).String(part)
	}
	return strings.Join(parts, "")
}

// テストデータとなるJSONを生成する関数
func genJson(paramSpecs []paramSpec) (string, error) {

	jsonBodyMap := map[string]any{}

	// パラメータ毎に型を判定し、テストデータを生成
	for _, param := range paramSpecs {

		// パラメータの型がstringの場合
		if param.Type == "string" {
			// exampleが設定されている場合はexampleを
			// 設定されていない場合はダミーデータを設定
			if param.Example != nil {
				jsonBodyMap[param.Name] = param.Example
			} else {
				jsonBodyMap[param.Name] = "dummy"
			}
			// パラメータの型がnumberの場合
		} else if param.Type == "number" {
			// 0を設定
			jsonBodyMap[param.Name] = 0
			// パラメータの型がそれ以外の場合
		} else {
			// 空文字列を設定
			jsonBodyMap[param.Name] = ""
		}
	}
	// JSONにシリアライズ
	jsonBodyObj, err := json.Marshal(jsonBodyMap)
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	// 文字列に変換して返す
	return string(jsonBodyObj), nil

}

func genItem(inputFileName string) (*[]pathSpec, error) {
	// パス毎の構造体を格納するスライスを定義
	var pathSpecs []pathSpec

	// OpenAPIのYAMLファイルを読み込み
	doc, err := openapi3.NewLoader().LoadFromFile(inputFileName)
	if err != nil {
		fmt.Println("Error:", err)
		return &[]pathSpec{}, err
	}

	// パス毎に処理
	for _, path := range doc.Paths.InMatchingOrder() {

		// それぞれのパスに対するメソッドの一覧を取得
		obj := doc.Paths.Find(path).Operations()

		// メソッド毎の構造体を格納するスライスを定義
		var baseSpecs []baseSpec

		// メソッド毎に処理
		for method, op := range obj {

			// クエリとボディに当たるパラメータ構造体を格納するスライスを定義
			var queries []paramSpec
			var bodies []paramSpec

			// 元データにクエリパラメータがある場合
			if op.Parameters != nil {
				for _, q := range op.Parameters {

					// クエリ毎にクエリパラメータ構造体を生成
					queries = append(queries, paramSpec{
						// フィールドの名前
						Name: q.Value.Name,
						// フィールドの型
						Type: q.Value.Schema.Value.Type,
						// フィールドのサンプル値
						Example: q.Value.Example,
					})
				}
			}

			// 元データにボディパラメータがある場合
			if op.RequestBody != nil {
				for name, b := range op.RequestBody.Value.Content["application/json"].Schema.Value.Properties {

					// ボディ毎にボディパラメータ構造体を生成
					bodies = append(bodies, paramSpec{
						// フィールドの名前
						Name: name,
						// フィールドの型
						Type: b.Value.Type,
						// フィールドのサンプル値
						Example: b.Value.Example,
					})
				}
			}

			// メソッド毎にメソッド構造体を生成して末尾に追加
			baseSpecs = append(baseSpecs, baseSpec{
				// メソッド名
				Method: method,
				// ボディパラメータ構造体のスライス
				Body: bodies,
				// クエリパラメータ構造体のスライス
				Params: queries,
			})

		}

		// パス毎にパス構造体を生成して末尾に追加
		pathSpecs = append(pathSpecs, pathSpec{
			// ディレクトリ名
			DirName: genDirName(path, "/"),
			// パス
			Path: path,
			// メソッド構造体のスライス
			Methods: baseSpecs,
		})

	}

	return &pathSpecs, nil
}

func renderTemplate(outputDir string, host string, pathSpecs []pathSpec) error {

	// ここからは先ほど生成した構造体を用いてテンプレートよりYAMLを作成
	// パス毎に処理
	for _, pathSpec := range pathSpecs {
		// メソッド毎に処理
		for _, methodItem := range pathSpec.Methods {

			// パスのパラメータを正規表現で処理して適した形に変換
			re := regexp.MustCompile(`{\s*([^}\s]*)\s*}`)
			modiPath := re.ReplaceAllString(pathSpec.Path, `{{ vars.req.query.$1 }}`)

			// 以下ディレクトリの作成
			// 0_base/パス/メソッド名/base.yml
			if err := os.MkdirAll(outputDir+"/0_base/"+pathSpec.DirName+"/"+methodItem.Method, 0777); err != nil {
				fmt.Println(err)
				return err
			}
			// 1_noAuth/パス/メソッド名/noAuth.yml
			if err := os.MkdirAll(outputDir+"/1_noAuth/"+pathSpec.DirName+"/"+methodItem.Method, 0777); err != nil {
				fmt.Println(err)
				return err
			}

			// 以下テンプレートの適用
			// 0_base/パス/メソッド名/base.ymlをレンダリングするための準備
			tmpl_base, err := template.New("base.yml.template").Delims("<<", ">>").ParseFiles("template/base.yml.template")
			if err != nil {
				fmt.Println(err)
				return err
			}
			// 0_base/パス/メソッド名/base.ymlファイルを作成
			fp_base, err := os.Create(outputDir + "/0_base/" + pathSpec.DirName + "/" + methodItem.Method + "/base.yml")
			if err != nil {
				fmt.Println(err)
				return err
			}
			// ファイルの後処理をdefer
			defer fp_base.Close()

			// ボディの形を適切に変換
			modiBody := ""
			// ボディがある場合
			if methodItem.Body != nil {
				// ボディの形を変換しmediatypeを付与
				modiBody = "application/json: \"{{ vars.req.body }}\""
				// ボディがなく、メソッドがGETでない場合
			} else if methodItem.Method != "get" {
				// ボディは空のまましmediatypeを付与
				modiBody = "application/json: []"
				// ボディがなく、メソッドがGETの場合
			} else {
				// ボディはnullを指定
				modiBody = "null"
			}

			// 0_base/パス/メソッド名/base.ymlをレンダリング
			err = tmpl_base.Execute(fp_base, map[string]any{
				"method": methodItem.Method,
				"path":   modiPath,
				"bodies": modiBody,
				"host":   host,
			})
			if err != nil {
				fmt.Println(err)
				return err
			}

			// 1_noAuth/パス/メソッド名/noAuth.ymlをレンダリングするための準備
			tmpl_noAuth, err := template.New("index.yml.template").Delims("<<", ">>").ParseFiles("template/index.yml.template")
			if err != nil {
				fmt.Println(err)
				return err
			}

			// 1_noAuth/パス/メソッド名/noAuth.ymlファイルを作成
			fp_noAuth, err := os.Create(outputDir + "/1_noAuth/" + pathSpec.DirName + "/" + methodItem.Method + "/base.yml")
			if err != nil {
				fmt.Println(err)
				return err
			}
			// ファイルの後処理をdefer
			defer fp_noAuth.Close()

			// 1_noAuth/パス/メソッド名/noAuth.ymlをレンダリング
			err = tmpl_noAuth.Execute(fp_noAuth, map[string]any{
				"desc":    "(" + methodItem.Method + ") " + pathSpec.Path + "のテスト",
				"method":  methodItem.Method,
				"dirname": pathSpec.DirName,
			})
			if err != nil {
				fmt.Println(err)
				return err
			}

			// Tomlの構成ファイルの読み込み
			var configToml config
			// config.tomlが存在する場合
			if _, err := os.Stat(outputDir + "/1_noAuth/" + pathSpec.DirName + "/" + methodItem.Method + "/config.toml"); !os.IsNotExist(err) {
				// config.tomlを読み込む
				_, err = toml.DecodeFile(outputDir+"/1_noAuth/"+pathSpec.DirName+"/"+methodItem.Method+"/config.toml", &configToml)
				if err != nil {
					fmt.Println(err)
					return err
				}
				// allowOverrideがfalseの場合
				if configToml.AllowOverride == false {
					// ループをコンティニューしJSONを生成しない
					continue
				}
			}

			// 1_noAuth/パス/メソッド名/data.jsonをレンダリングするための準備
			tmpl_data, err := template.New("data.json.template").Delims("<<", ">>").ParseFiles("template/data.json.template")
			if err != nil {
				fmt.Println(err)
				return err
			}
			// 1_noAuth/パス/メソッド名/data.jsonファイルを作成
			fp_data, err := os.Create(outputDir + "/1_noAuth/" + pathSpec.DirName + "/" + methodItem.Method + "/data.json")
			if err != nil {
				fmt.Println(err)
				return err
			}
			// ファイルの後処理をdefer
			defer fp_data.Close()

			// ボディのJSONを生成
			jsonBody, err := genJson(methodItem.Body)
			if err != nil {
				return err
			}
			// クエリのJSONを生成
			jsonQuery, err := genJson(methodItem.Params)
			if err != nil {
				return err
			}

			// 1_noAuth/パス/メソッド名/data.jsonをレンダリング
			if err = tmpl_data.Execute(fp_data, map[string]any{
				"jsonBody":  jsonBody,
				"jsonQuery": jsonQuery,
			}); err != nil {
				fmt.Println(err)
				return err
			}

		}
	}

	return nil
}
