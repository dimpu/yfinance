package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	yahoo "github.com/dimpu/yfinance"
	"gopkg.in/yaml.v3"
)

var useYAML bool

var rootCmd = &cobra.Command{
	Use:   "yf",
	Short: "Yahoo Finance CLI",
	Long:  "Command-line interface for the Yahoo Finance API.",
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&useYAML, "yaml", false, "output in YAML format")
}

func newClient() *yahoo.Client {
	return yahoo.NewClient(nil)
}

func printResult(v interface{}) error {
	if useYAML {
		out, err := yaml.Marshal(v)
		if err != nil {
			return fmt.Errorf("yaml marshal: %w", err)
		}
		fmt.Print(string(out))
		return nil
	}
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	fmt.Println(string(out))
	return nil
}

func fatal(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
	} else {
		fmt.Fprintf(os.Stderr, "%s\n", msg)
	}
	os.Exit(1)
}
