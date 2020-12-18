package cmd

import (
	"os"

	"github.com/semaphoreci/artifact/pkg/gcs"
	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	"github.com/semaphoreci/artifact/pkg/util/log"
	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Stores a file or directory in the storage for later use",
	Long: `You may store project, workflow or job related files, that you can use
while the rest of the semaphore process, or after it.`,
}

func runPushForCategory(cmd *cobra.Command, args []string, category, catID,
	expireDefault string) (string, string) {
	err := pathutil.InitPathID(category, catID)
	errutil.Check(err)
	src := args[0]

	dst, err := cmd.Flags().GetString("destination")
	errutil.Check(err)

	force, err := cmd.Flags().GetBool("force")
	errutil.Check(err)

	expireIn, err := cmd.Flags().GetString("expire-in")
	errutil.Check(err)
	if len(expireIn) == 0 {
		expireIn = expireDefault
	}

	dst, src = gcs.PushPaths(dst, src)
	if ok := gcs.PushGCS(dst, src, expireIn, force); !ok {
		os.Exit(1) // error already logged
	}
	return dst, src
}

// PushJobCmd is the subcommand for "artifact push job ..."
var PushJobCmd = &cobra.Command{
	Use:   "job [SOURCE PATH]",
	Short: "Uploads a job file or directory to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("job-id")
		errutil.Check(err)
		dst, src := runPushForCategory(cmd, args, pathutil.JOB, catID,
			viper.GetString("JobArtifactsExpire"))
		log.Info("successful push for current job", zap.String("source", src),
			zap.String("destination", dst))
	},
}

// PushWorkflowCmd is the subcommand for "artifact push workflow ..."
var PushWorkflowCmd = &cobra.Command{
	Use:   "workflow [SOURCE PATH]",
	Short: "Uploads a workflow or directory file to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("workflow-id")
		errutil.Check(err)
		dst, src := runPushForCategory(cmd, args, pathutil.WORKFLOW, catID,
			viper.GetString("WorkflowArtifactsExpire"))
		log.Info("successful push for current workflow", zap.String("source", src),
			zap.String("destination", dst))
	},
}

// PushProjectCmd is the subcommand for "artifact push project ..."
var PushProjectCmd = &cobra.Command{
	Use:   "project [SOURCE PATH]",
	Short: "Upload a project file or directory to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("project-id")
		errutil.Check(err)
		dst, src := runPushForCategory(cmd, args, pathutil.PROJECT, catID,
			viper.GetString("ProjectArtifactsExpire"))
		log.Info("successful push for current project", zap.String("source", src),
			zap.String("destination", dst))
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.AddCommand(PushJobCmd)
	pushCmd.AddCommand(PushWorkflowCmd)
	pushCmd.AddCommand(PushProjectCmd)

	desc := "rename the file while uploading"
	PushJobCmd.Flags().StringP("destination", "d", "", desc)
	PushWorkflowCmd.Flags().StringP("destination", "d", "", desc)
	PushProjectCmd.Flags().StringP("destination", "d", "", desc)

	desc = "force overwrite"
	PushJobCmd.Flags().BoolP("force", "f", false, desc)
	PushWorkflowCmd.Flags().BoolP("force", "f", false, desc)
	PushProjectCmd.Flags().BoolP("force", "f", false, desc)

	desc = `Removes the files after the given amount of time.
Nd for N days
Nw for N weeks
Nm for N months
Ny for N years
`
	PushJobCmd.Flags().StringP("expire-in", "e", "", desc)
	PushWorkflowCmd.Flags().StringP("expire-in", "e", "", desc)
	PushProjectCmd.Flags().StringP("expire-in", "e", "", desc)

	PushJobCmd.Flags().StringP("job-id", "j", "", "set explicit job id")
	PushWorkflowCmd.Flags().StringP("workflow-id", "w", "", "set explicit workflow id")
	PushProjectCmd.Flags().StringP("project-id", "p", "", "set explicit project id")
}
