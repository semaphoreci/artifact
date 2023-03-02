package cmd

import (
	errutil "github.com/semaphoreci/artifact/pkg/errors"
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

func runYankForCategory(cmd *cobra.Command, args []string, resolver *files.PathResolver) (*files.ResolvedPath, error) {
	hubClient, err := hub.NewClient()
	errutil.Check(err)

	// The yank operation does not have a destination override
	paths, err := resolver.Resolve(files.OperationYank, args[0], "")
	errutil.Check(err)

	return paths, storage.Yank(hubClient, paths.Source, verbose)
}

func NewYankJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job [PATH]",
		Short: "Deletes a job file or directory from the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			jobId, err := cmd.Flags().GetString("job-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeJob, jobId)
			errutil.Check(err)

			paths, err := runYankForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error yanking artifact: %v\n", err)
				log.Error("Please check if the artifact you are trying to yank exists.\n")
				errutil.Exit(1)
				return
			}

			log.Infof("Successfully yanked '%s' from current job artifacts.\n", paths.Source)
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
			workflowId, err := cmd.Flags().GetString("workflow-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeWorkflow, workflowId)
			errutil.Check(err)

			paths, err := runYankForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error yanking artifact: %v\n", err)
				log.Error("Please check if the artifact you are trying to yank exists.\n")
				errutil.Exit(1)
				return
			}

			log.Infof("Successfully yanked '%s' from current workflow artifacts.\n", paths.Source)
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
			projectId, err := cmd.Flags().GetString("project-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeProject, projectId)
			errutil.Check(err)

			paths, err := runYankForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error yanking artifact: %v\n", err)
				log.Error("Please check if the artifact you are trying to yank exists.\n")
				errutil.Exit(1)
				return
			}

			log.Infof("Successfully yanked '%s' from current project artifacts.\n", paths.Source)
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
