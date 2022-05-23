package cmd

import (
	"path/filepath"

	errutil "github.com/semaphoreci/artifact/pkg/err"
	"github.com/semaphoreci/artifact/pkg/files"
	"github.com/semaphoreci/artifact/pkg/hub"
	"github.com/semaphoreci/artifact/pkg/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var yankCmd = &cobra.Command{
	Use:     "yank",
	Aliases: []string{"delete"},
	Short:   "Deletes a file or directory from the storage you pushed earlier",
	Long: `You may store files project, workflow or job related files with
artifact push. With artifact yank you can delete them if you
don't need them any more.`,
}

func runYankForCategory(cmd *cobra.Command, args []string, category, catID string) (string, error) {
	hubClient, err := hub.NewClient()
	errutil.Check(err)

	err = files.InitPathID(category, catID)
	errutil.Check(err)
	name := args[0]

	name = files.YankPath(filepath.ToSlash(name))
	return name, storage.Yank(hubClient, name)
}

func NewYankJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job [PATH]",
		Short: "Deletes a job file or directory from the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			catID, err := cmd.Flags().GetString("job-id")
			errutil.Check(err)

			name, err := runYankForCategory(cmd, args, files.JOB, catID)
			if err != nil {
				log.Errorf("Error yanking artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Infof("Successfully yanked '%s' from current job artifacts.\n", name)
		},
	}

	cmd.Flags().StringP("job-id", "j", "", "set explicit job id")
	return cmd
}

func NewYankWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow [PATH]",
		Short: "Deletes a workflow file or directory from the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			catID, err := cmd.Flags().GetString("workflow-id")
			errutil.Check(err)

			name, err := runYankForCategory(cmd, args, files.WORKFLOW, catID)
			if err != nil {
				log.Errorf("Error yanking artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Infof("Successfully yanked '%s' from current workflow artifacts.\n", name)
		},
	}

	cmd.Flags().StringP("workflow-id", "w", "", "set explicit workflow id")
	return cmd
}

func NewYankProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [PATH]",
		Short: "Deletes a project file or directory from the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			catID, err := cmd.Flags().GetString("project-id")
			errutil.Check(err)

			name, err := runYankForCategory(cmd, args, files.PROJECT, catID)
			if err != nil {
				log.Errorf("Error yanking artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Infof("Successfully yanked '%s' from current project artifacts.\n", name)
		},
	}

	cmd.Flags().StringP("project-id", "p", "", "set explicit project id")
	return cmd
}

func init() {
	rootCmd.AddCommand(yankCmd)
	yankCmd.AddCommand(NewYankJobCmd())
	yankCmd.AddCommand(NewYankWorkflowCmd())
	yankCmd.AddCommand(NewYankProjectCmd())
}
