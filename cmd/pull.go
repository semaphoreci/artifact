package cmd

import (
	"github.com/semaphoreci/artifact/pkg/gcs"
	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
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

func runPullForCategory(cmd *cobra.Command, args []string, category, catID string) (string, string) {
	pathutil.InitPathID(category, catID)
	src := args[0]

	dst, err := cmd.Flags().GetString("destination")
	errutil.Check(err)

	force, err := cmd.Flags().GetBool("force")
	errutil.Check(err)

	dst, src = gcs.PullPaths(dst, src)
	err = gcs.PullGCS(dst, src, force)
	errutil.Check(err)
	return dst, src
}

// PullJobCmd is the subcommand for "artifact pull job ..."
var PullJobCmd = &cobra.Command{
	Use:   "job [SOURCE PATH]",
	Short: "Downloads a job file or directory from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("job-id")
		errutil.Check(err)
		dst, src := runPullForCategory(cmd, args, pathutil.JOB, catID)
		errutil.Info("'%s' pulled to '%s' for current job.\n", src, dst)
	},
}

// PullWorkflowCmd is the subcommand for "artifact pull workflow ..."
var PullWorkflowCmd = &cobra.Command{
	Use:   "workflow [SOURCE PATH]",
	Short: "Downloads a workflow file or directory from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("workflow-id")
		errutil.Check(err)
		dst, src := runPullForCategory(cmd, args, pathutil.WORKFLOW, catID)
		errutil.Info("'%s' pulled to '%s' for current workflow.\n", src, dst)
	},
}

// PullProjectCmd is the subcommand for "artifact pull project ..."
var PullProjectCmd = &cobra.Command{
	Use:   "project [SOURCE PATH]",
	Short: "Downloads a project file or directory from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		dst, src := runPullForCategory(cmd, args, pathutil.PROJECT, "")
		errutil.Info("'%s' pulled to '%s' for current project.\n", src, dst)
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

	PullJobCmd.Flags().StringP("job-id", "j", "", "set explicit job id")
	PullWorkflowCmd.Flags().StringP("workflow-id", "w", "", "set explicit workflow id")
}
