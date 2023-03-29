/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "openapi2runn",
	Short: "Create test scenario files for runn based on OpenAPI documentation",
	Long: `The process of creating test scenario files for runn from OpenAPI documentation involves creating scenarios for each API method endpoint,
and placing data in JSON files in the same directory as the scenarios. By increasing the number of arrays in the JSON file, multiple test data can be included in the scenarios.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
