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
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("yank called")
	// },
}

// YankJobCmd is the subcommand for "artifact yank job ..."
var YankJobCmd = &cobra.Command{
	Use:   "job [PATH]",
	Short: "Deletes a job file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		filename, err := yankFileGCS(utils.JOB, filename)
		utils.Check(err)
		fmt.Printf("File '%s' deleted for current job.\n", filename)
	},
}

// YankWorkflowCmd is the subcommand for "artifact yank workflow ..."
var YankWorkflowCmd = &cobra.Command{
	Use:   "workflow [PATH]",
	Short: "Deletes a workflow file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		filename, err := yankFileGCS(utils.WORKFLOW, filename)
		utils.Check(err)
		fmt.Printf("File '%s' deleted for current workflow.\n", filename)
	},
}

// YankProjectCmd is the subcommand for "artifact yank project ..."
var YankProjectCmd = &cobra.Command{
	Use:   "project [PATH]",
	Short: "Deletes a project file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		filename := args[0]

		filename, err := yankFileGCS(utils.PROJECT, filename)
		utils.Check(err)
		fmt.Printf("File '%s' deleted for current project.\n", filename)
	},
}

func init() {
	rootCmd.AddCommand(yankCmd)

	yankCmd.AddCommand(YankJobCmd)
	yankCmd.AddCommand(YankWorkflowCmd)
	yankCmd.AddCommand(YankProjectCmd)
}
