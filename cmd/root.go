package cmd

import (
	"fmt"
	"os"

	"github.com/semaphoreci/artifact/cmd/utils"
	"github.com/spf13/cobra"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Helps the semaphore to store files through the process",
	Long:  "",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	err := initGCS()
	utils.Check(err)
}
