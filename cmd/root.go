package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/hcvelo/hcvelo/cmd/strava"
)

var (
	cfgDir string

	rootCmd = &cobra.Command{
		Use:   "hcvelo",
		Short: "hcvelo is a cli for the Holmes Chapel Velo cycling club",
		Long:  "hcvelo is a cli for the Holmes Chapel Velo cycling club",
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				return
			}
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
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgDir, "configDir", "", "config directory (default is $HOME/.hcvelo)")

	rootCmd.AddCommand(strava.Cmd)
}

func initConfig() {
	if cfgDir == "" {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		cfgDir = filepath.Join(home, ".hcvelo")
	}
}
