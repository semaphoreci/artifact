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

// pullCmd represents the pull command
var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Downloads a file or directory from the storage you pushed earlier",
	Long: `You may store files project, workflow or job related files with
artifact push. With artifact pull you can download them to the current directory
to use them in a later phase, debug, or getting the results.`,
}

func runPullForCategory(cmd *cobra.Command, args []string, category, catID string) (string, string, error) {
	hubClient, err := hub.NewClient()
	errutil.Check(err)

	err = files.InitPathID(category, catID)
	errutil.Check(err)
	src := args[0]

	dst, err := cmd.Flags().GetString("destination")
	errutil.Check(err)

	force, err := cmd.Flags().GetBool("force")
	errutil.Check(err)

	dst, src = files.PullPaths(filepath.ToSlash(dst), filepath.ToSlash(src))
	return dst, src, storage.Pull(hubClient, dst, src, force)
}

func NewPullJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job [SOURCE PATH]",
		Short: "Downloads a job file or directory from the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			catID, err := cmd.Flags().GetString("job-id")
			errutil.Check(err)

			dst, src, err := runPullForCategory(cmd, args, files.JOB, catID)
			if err != nil {
				log.Errorf("Error pulling artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pulled artifact for current job.\n")
			log.Infof("> Source: '%s'.\n", src)
			log.Infof("> Destination: '%s'.\n", dst)
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
			catID, err := cmd.Flags().GetString("workflow-id")
			errutil.Check(err)
			dst, src, err := runPullForCategory(cmd, args, files.WORKFLOW, catID)
			if err != nil {
				log.Errorf("Error pulling artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pulled artifact for current workflow.\n")
			log.Infof("> Source: '%s'.\n", src)
			log.Infof("> Destination: '%s'.\n", dst)
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
			dst, src, err := runPullForCategory(cmd, args, files.PROJECT, "")
			if err != nil {
				log.Errorf("Error pulling artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pulled artifact for current project.\n")
			log.Infof("> Source: '%s'.\n", src)
			log.Infof("> Destination: '%s'.\n", dst)
		},
	}

	cmd.Flags().StringP("destination", "d", "", "rename the file while uploading")
	cmd.Flags().BoolP("force", "f", false, "force overwrite")
	cmd.Flags().StringP("project-id", "p", "", "set explicit project id")
	return cmd
}

func init() {
	rootCmd.AddCommand(pullCmd)

	pullJobCmd := NewPullJobCmd()
	pullWorkflowCmd := NewPullWorkflowCmd()
	pullProjectCmd := NewPullProjectCmd()

	pullCmd.AddCommand(pullJobCmd)
	pullCmd.AddCommand(pullWorkflowCmd)
	pullCmd.AddCommand(pullProjectCmd)
}
