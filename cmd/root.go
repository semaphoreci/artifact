package cmd

import (
	homedir "github.com/mitchellh/go-homedir"
	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile string
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "artifact",
	Short: "Semaphore 2.0 Artifact CLI",
	Long:  "",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		errutil.Init(verbose)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	errutil.Check(err)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.artifact.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose logging")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetDefault("ProjectArtifactsExpire", "never")
	viper.SetDefault("WorkflowArtifactsExpire", "never")
	viper.SetDefault("JobArtifactsExpire", "never")

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		errutil.Check(err)

		// Search config in home directory with name ".artifact" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".artifact")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		errutil.L.Debug("Using config file", zap.String("filename", viper.ConfigFileUsed()))
	}
}
