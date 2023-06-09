/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// genCmd represents the gen command
var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Create test scenario files for runn based on OpenAPI documentation",
	Long: `The process of creating test scenario files for runn from OpenAPI documentation involves creating scenarios for each API method endpoint,
and placing data in JSON files in the same directory as the scenarios. By increasing the number of arrays in the JSON file, multiple test data can be included in the scenarios.`,
	PreRun: func(cmd *cobra.Command, args []string) {
		fmt.Println(`
_____  _____  _____  _____  _____  _____  ___  _____  _____  __ __  _____  _____ 
/  _  \/  _  \/   __\/  _  \/  _  \/  _  \/___\<___  \/  _  \/  |  \/  _  \/  _  \
|  |  ||   __/|   __||  |  ||  _  ||   __/|   | /  __/|  _  <|  |  ||  |  ||  |  |
\_____/\__/   \_____/\__|__/\__|__/\__/   \___/<_____|\__|\_/\_____/\__|__/\__|__/
`)
	},

	Run: func(cmd *cobra.Command, args []string) {

		// OpenAPIのYAMLファイルを読み込みしてオブジェクトを生成
		flags := *cmd.Flags()

		// フラグから入力ファイル名を取得
		input, err := flags.GetString("input")
		if err != nil {
			fmt.Println(err)
			input = "openapi.yml"
		}

		// フラグから出力ディレクトリ名を取得
		output, err := flags.GetString("output")
		if err != nil {
			fmt.Println(err)
			output = "output"
		}

		// フラグから出力ディレクトリ名を取得
		host, err := flags.GetString("server")
		if err != nil {
			fmt.Println(err)
			output = "http://localhost:8080"
		}

		// OpenAPIのYAMLファイルを読み込みしてオブジェクトを生成
		pathSpecs, err := genItem(input)
		if err != nil {
			fmt.Println(err)
			return
		}

		// テンプレートをレンダリング
		err = renderTemplate(output, host, *pathSpecs)
		if err != nil {
			fmt.Println(err)
			return
		}
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
	genCmd.Flags().StringP("input", "i", "", "Input file name")
	genCmd.Flags().StringP("output", "o", "", "Output dir name")
	genCmd.Flags().StringP("server", "s", "", "Host server")

	genCmd.MarkFlagRequired("input")
	genCmd.MarkFlagRequired("output")
	genCmd.MarkFlagRequired("server")

}
