package cmd

import (
	"github.com/semaphoreci/artifact/pkg/gcs"
	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
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

func runYankForCategory(cmd *cobra.Command, args []string, category, catID string) string {
	pathutil.InitPathID(category, catID)
	name := args[0]

	name = gcs.YankPath(name)
	err := gcs.YankGCS(name)
	errutil.Check(err)
	return name
}

// YankJobCmd is the subcommand for "artifact yank job ..."
var YankJobCmd = &cobra.Command{
	Use:   "job [PATH]",
	Short: "Deletes a job file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("job-id")
		errutil.Check(err)
		name := runYankForCategory(cmd, args, pathutil.JOB, catID)
		errutil.Info("File '%s' deleted for current job.\n", name)
	},
}

// YankWorkflowCmd is the subcommand for "artifact yank workflow ..."
var YankWorkflowCmd = &cobra.Command{
	Use:   "workflow [PATH]",
	Short: "Deletes a workflow file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("workflow-id")
		errutil.Check(err)
		name := runYankForCategory(cmd, args, pathutil.WORKFLOW, catID)
		errutil.Info("File '%s' deleted for current workflow.\n", name)
	},
}

// YankProjectCmd is the subcommand for "artifact yank project ..."
var YankProjectCmd = &cobra.Command{
	Use:   "project [PATH]",
	Short: "Deletes a project file from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		name := runYankForCategory(cmd, args, pathutil.PROJECT, "")
		errutil.Info("File '%s' deleted for current project.\n", name)
	},
}

func init() {
	rootCmd.AddCommand(yankCmd)

	yankCmd.AddCommand(YankJobCmd)
	yankCmd.AddCommand(YankWorkflowCmd)
	yankCmd.AddCommand(YankProjectCmd)

	YankJobCmd.Flags().StringP("job-id", "j", "", "set explicit job id")
	YankWorkflowCmd.Flags().StringP("workflow-id", "w", "", "set explicit workflow id")
}
