package cmd

import (
	"fmt"

	"github.com/semaphoreci/artifact/cmd/utils"
	"github.com/spf13/cobra"
)

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Downloads a file or directory from the storage you pushed earlier",
	Long: `You may store files project, workflow or job related files with
artifact push. With artifact pull you can download them to the current directory
to use them in a later phase, debug, or getting the results.`,
}

func runPullForCategory(cmd *cobra.Command, args []string, category string) (string, string) {
	src := args[0]

	dst, err := cmd.Flags().GetString("destination")
	utils.Check(err)

	force, err := cmd.Flags().GetBool("force")
	utils.Check(err)

	dst, src = pullPaths(category, dst, src)
	err = pullGCS(dst, src, force)
	utils.Check(err)
	return dst, src
}

// PullJobCmd is the subcommand for "artifact pull job ..."
var PullJobCmd = &cobra.Command{
	Use:   "job [SOURCE PATH]",
	Short: "Download a job file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		dst, src := runPullForCategory(cmd, args, utils.JOB)
		fmt.Printf("File '%s' pulled to '%s' for current job.\n", src, dst)
	},
}

// PullWorkflowCmd is the subcommand for "artifact pull workflow ..."
var PullWorkflowCmd = &cobra.Command{
	Use:   "workflow [SOURCE PATH]",
	Short: "Download a workflow file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		dst, src := runPullForCategory(cmd, args, utils.WORKFLOW)
		fmt.Printf("File '%s' pulled to '%s' for current workflow.\n", src, dst)
	},
}

// PullProjectCmd is the subcommand for "artifact pull project ..."
var PullProjectCmd = &cobra.Command{
	Use:   "project [SOURCE PATH]",
	Short: "Download a project file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		dst, src := runPullForCategory(cmd, args, utils.PROJECT)
		fmt.Printf("File '%s' pulled to '%s' for current project.\n", src, dst)
	},
}

func init() {
	rootCmd.AddCommand(pullCmd)

	pullCmd.AddCommand(PullJobCmd)
	pullCmd.AddCommand(PullWorkflowCmd)
	pullCmd.AddCommand(PullProjectCmd)

	desc := "rename the file while uploading"
	PullJobCmd.Flags().StringP("destination", "d", "", desc)
	PullWorkflowCmd.Flags().StringP("destination", "d", "", desc)
	PullProjectCmd.Flags().StringP("destination", "d", "", desc)

	desc = "force overwrite"
	PullJobCmd.Flags().BoolP("force", "f", false, desc)
	PullWorkflowCmd.Flags().BoolP("force", "f", false, desc)
	PullProjectCmd.Flags().BoolP("force", "f", false, desc)
}
