package cmd

import (
	"fmt"

	"github.com/semaphoreci/artifact/internal"
	"github.com/semaphoreci/artifact/pkg/gcs"
	"github.com/semaphoreci/artifact/pkg/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
	utils.InitPathID(category, catID)
	src := args[0]

	dst, err := cmd.Flags().GetString("destination")
	internal.Check(err)

	force, err := cmd.Flags().GetBool("force")
	internal.Check(err)

	expireIn, err := cmd.Flags().GetString("expire-in")
	internal.Check(err)
	if len(expireIn) == 0 {
		expireIn = expireDefault
	}

	dst, src = gcs.PushPaths(dst, src)
	_, err = gcs.PushGCS(dst, src, expireIn, force)
	internal.Check(err)
	return dst, src
}

// PushJobCmd is the subcommand for "artifact push job ..."
var PushJobCmd = &cobra.Command{
	Use:   "job [SOURCE PATH]",
	Short: "Upload a job file to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("job-id")
		internal.Check(err)
		dst, src := runPushForCategory(cmd, args, utils.JOB, catID,
			viper.GetString("JobArtifactsExpire"))
		fmt.Printf("File '%s' pushed to '%s' for current job.\n", src, dst)
	},
}

// PushWorkflowCmd is the subcommand for "artifact push workflow ..."
var PushWorkflowCmd = &cobra.Command{
	Use:   "workflow [SOURCE PATH]",
	Short: "Upload a workflow file to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		catID, err := cmd.Flags().GetString("workflow-id")
		internal.Check(err)
		dst, src := runPushForCategory(cmd, args, utils.WORKFLOW, catID,
			viper.GetString("WorkflowArtifactsExpire"))
		fmt.Printf("File '%s' pushed to '%s' for current workflow.\n", src, dst)
	},
}

// PushProjectCmd is the subcommand for "artifact push project ..."
var PushProjectCmd = &cobra.Command{
	Use:   "project [SOURCE PATH]",
	Short: "Upload a project file to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		dst, src := runPushForCategory(cmd, args, utils.PROJECT, "",
			viper.GetString("ProjectArtifactsExpire"))
		fmt.Printf("File '%s' pushed to '%s' for current project.\n", src, dst)
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
just integer (number of seconds)
Nh for N hours
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
}
