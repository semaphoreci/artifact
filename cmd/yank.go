package cmd

import (
	"os"
	"path/filepath"

	errutil "github.com/semaphoreci/artifact/pkg/err"
	"github.com/semaphoreci/artifact/pkg/files"
	"github.com/semaphoreci/artifact/pkg/hub"
	"github.com/semaphoreci/artifact/pkg/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// yankCmd represents the yank command
var yankCmd = &cobra.Command{
	Use:     "yank",
	Aliases: []string{"delete"},
	Short:   "Deletes a file or directory from the storage you pushed earlier",
	Long: `You may store files project, workflow or job related files with
artifact push. With artifact yank you can delete them if you
don't need them any more.`,
}

func runYankForCategory(cmd *cobra.Command, args []string, category, catID string) string {
	hubClient, err := hub.NewClient()
	errutil.Check(err)

	err = files.InitPathID(category, catID)
	errutil.Check(err)
	name := args[0]

	name = files.YankPath(filepath.ToSlash(name))
	err = storage.Yank(hubClient, name)
	if err != nil {
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
		name := runYankForCategory(cmd, args, files.JOB, catID)
		log.Infof("Successfully yanked '%s' from current job artifacts.\n", name)
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
		name := runYankForCategory(cmd, args, files.WORKFLOW, catID)
		log.Infof("Successfully yanked '%s' from current workflow artifacts.\n", name)
	},
}

// YankProjectCmd is the subcommand for "artifact yank project ..."
var YankProjectCmd = &cobra.Command{
	Use:   "project [PATH]",
	Short: "Deletes a project file or directory from the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		name := runYankForCategory(cmd, args, files.PROJECT, "")
		log.Infof("Successfully yanked '%s' from current project artifacts.\n", name)
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
