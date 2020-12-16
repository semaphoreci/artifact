package cmd

import (
	"os"

	"github.com/semaphoreci/artifact/pkg/gcs"
	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	"github.com/semaphoreci/artifact/pkg/util/log"
	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
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
	if ok := gcs.YankGCS(name); !ok {
		os.Exit(1) // error already logged
	}
	return name
}

// YankJobCmd is the subcommand for "artifact yank job ..."
var YankJobCmd = &cobra.Command{
	Use:   "job [PATH]",
	Short: "Deletes a job file or directory from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("job-id")
		errutil.Check(err)
		name := runYankForCategory(cmd, args, pathutil.JOB, catID)
		log.Info("successful yank for current job", zap.String("name", name))
	},
}

// YankWorkflowCmd is the subcommand for "artifact yank workflow ..."
var YankWorkflowCmd = &cobra.Command{
	Use:   "workflow [PATH]",
	Short: "Deletes a workflow file or directory from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("workflow-id")
		errutil.Check(err)
		name := runYankForCategory(cmd, args, pathutil.WORKFLOW, catID)
		log.Info("successful yank for current workflow", zap.String("name", name))
	},
}

// YankProjectCmd is the subcommand for "artifact yank project ..."
var YankProjectCmd = &cobra.Command{
	Use:   "project [PATH]",
	Short: "Deletes a project file or directory from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		name := runYankForCategory(cmd, args, pathutil.PROJECT, "")
		log.Info("successful yank for current project", zap.String("name", name))
	},
}

func init() {
	rootCmd.AddCommand(yankCmd)

	yankCmd.AddCommand(YankJobCmd)
	yankCmd.AddCommand(YankWorkflowCmd)
	yankCmd.AddCommand(YankProjectCmd)

	YankJobCmd.Flags().StringP("job-id", "j", "", "set explicit job id")
	YankWorkflowCmd.Flags().StringP("workflow-id", "w", "", "set explicit workflow id")
	YankProjectCmd.Flags().StringP("project-id", "p", "", "set explicit project id")
}
