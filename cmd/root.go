package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/hcvelo/hcvelo/cmd/strava"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{
		Use:   "hcvelo",
		Short: "hcvelo is a cli for the Holmes Chapel Velo cycling club",
		Long:  "hcvelo is a cli for the Holmes Chapel Velo cycling club",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}
)

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(strava.Cmd)
}
