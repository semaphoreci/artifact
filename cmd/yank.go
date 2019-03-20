package cmd

import (
	"fmt"

	"github.com/semaphoreci/artifact/cmd/utils"
	"github.com/spf13/cobra"
)

// yankCmd represents the yank command
var yankCmd = &cobra.Command{
	Use:   "yank",
	Short: "Deletes a file or directory from the storage you pushed earlier",
	Long: `You may store files project, workflow or job related files with
artifact push. With artifact yank you can delete them if you
don't need them any more.`,
}

func runYankForCategory(cmd *cobra.Command, args []string, category string) string {
	name := args[0]

	name = yankPath(category, name)
	err := yankGCS(name)
	utils.Check(err)
	return name
}

// YankJobCmd is the subcommand for "artifact yank job ..."
var YankJobCmd = &cobra.Command{
	Use:   "job [PATH]",
	Short: "Deletes a job file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		name := runYankForCategory(cmd, args, utils.JOB)
		fmt.Printf("File '%s' deleted for current job.\n", name)
	},
}

// YankWorkflowCmd is the subcommand for "artifact yank workflow ..."
var YankWorkflowCmd = &cobra.Command{
	Use:   "workflow [PATH]",
	Short: "Deletes a workflow file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		name := runYankForCategory(cmd, args, utils.WORKFLOW)
		fmt.Printf("File '%s' deleted for current workflow.\n", name)
	},
}

// YankProjectCmd is the subcommand for "artifact yank project ..."
var YankProjectCmd = &cobra.Command{
	Use:   "project [PATH]",
	Short: "Deletes a project file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		name := runYankForCategory(cmd, args, utils.PROJECT)
		fmt.Printf("File '%s' deleted for current project.\n", name)
	},
}

func init() {
	rootCmd.AddCommand(yankCmd)

	yankCmd.AddCommand(YankJobCmd)
	yankCmd.AddCommand(YankWorkflowCmd)
	yankCmd.AddCommand(YankProjectCmd)
}
