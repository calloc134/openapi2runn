package main

import (
	"fmt"
)

func main() {
	// OpenAPIのYAMLファイルを読み込みしてオブジェクトを生成
	pathSpecs, err := genItem("openapi.yaml")
	if err != nil {
		fmt.Println(err)
	}

	// テンプレートをレンダリング
	err = renderTemplate("output", *pathSpecs)
	if err != nil {
		fmt.Println(err)
	}

}
