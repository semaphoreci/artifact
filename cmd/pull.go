package cmd

import (
	errutil "github.com/semaphoreci/artifact/pkg/errors"
	"github.com/semaphoreci/artifact/pkg/files"
	"github.com/semaphoreci/artifact/pkg/hub"
	"github.com/semaphoreci/artifact/pkg/storage"
	log "github.com/sirupsen/logrus"
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

// Maybe use an api.Artifact?
func runPullForCategory(cmd *cobra.Command, args []string, resolver *files.PathResolver) (*files.ResolvedPath, error) {
	destinationOverride, err := cmd.Flags().GetString("destination")
	errutil.Check(err)

	force, err := cmd.Flags().GetBool("force")
	errutil.Check(err)

	hubClient, err := hub.NewClient()
	errutil.Check(err)

	return storage.Pull(hubClient, resolver, storage.PullOptions{
		SourcePath:          args[0],
		DestinationOverride: destinationOverride,
		Force:               force,
	})
}

func NewPullJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job [SOURCE PATH]",
		Short: "Downloads a job file or directory from the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			jobId, err := cmd.Flags().GetString("job-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeJob, jobId)
			errutil.Check(err)

			paths, err := runPullForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error pulling artifact: %v\n", err)
				log.Error("Please check if the artifact you are trying to pull exists.\n")
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pulled artifact for current job.\n")
			log.Infof("* Remote source: '%s'.\n", paths.Source)
			log.Infof("* Local destination: '%s'.\n", paths.Destination)
		},
	}

	cmd.Flags().StringP("destination", "d", "", "rename the file while uploading")
	cmd.Flags().BoolP("force", "f", false, "force overwrite")
	cmd.Flags().StringP("job-id", "j", "", "set explicit job id")
	return cmd
}

func NewPullWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow [SOURCE PATH]",
		Short: "Downloads a workflow file or directory from the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			workflowId, err := cmd.Flags().GetString("workflow-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeWorkflow, workflowId)
			errutil.Check(err)

			paths, err := runPullForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error pulling artifact: %v\n", err)
				log.Error("Please check if the artifact you are trying to pull exists.\n")
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pulled artifact for current workflow.\n")
			log.Infof("* Remote source: '%s'.\n", paths.Source)
			log.Infof("* Local destination: '%s'.\n", paths.Destination)
		},
	}

	cmd.Flags().StringP("destination", "d", "", "rename the file while uploading")
	cmd.Flags().BoolP("force", "f", false, "force overwrite")
	cmd.Flags().StringP("workflow-id", "w", "", "set explicit workflow id")
	return cmd
}

func NewPullProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [SOURCE PATH]",
		Short: "Downloads a project file or directory from the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			projectId, err := cmd.Flags().GetString("project-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeProject, projectId)
			errutil.Check(err)

			paths, err := runPullForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error pulling artifact: %v\n", err)
				log.Error("Please check if the artifact you are trying to pull exists.\n")
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pulled artifact for current project.\n")
			log.Infof("* Remote source: '%s'.\n", paths.Source)
			log.Infof("* Local destination: '%s'.\n", paths.Destination)
		},
	}

	cmd.Flags().StringP("destination", "d", "", "rename the file while uploading")
	cmd.Flags().BoolP("force", "f", false, "force overwrite")
	cmd.Flags().StringP("project-id", "p", "", "set explicit project id")
	return cmd
}

func init() {
	rootCmd.AddCommand(pullCmd)
	pullCmd.AddCommand(NewPullJobCmd())
	pullCmd.AddCommand(NewPullWorkflowCmd())
	pullCmd.AddCommand(NewPullProjectCmd())
}
