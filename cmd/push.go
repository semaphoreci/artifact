package cmd

import (
	"fmt"

	"github.com/semaphoreci/artifact/cmd/utils"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Stores a file or directory in the storage for later use",
	Long: `You may store project, workflow or job related files, that you can use
while the rest of the semaphore process, or after it.`,
	// Run: func(cmd *cobra.Command, args []string) {
	// 	fmt.Println("push called")
	// },
}

var PushJobCmd = &cobra.Command{
	Use:   "job [SOURCE PATH]",
	Short: "Upload a job file to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		src := args[0]

		dst, err := cmd.Flags().GetString("destination")
		utils.Check(err)

		expireIn, err := cmd.Flags().GetString("expire-in")
		utils.Check(err)

		dst, src, err = pushFileGCS(utils.JOB, dst, src, expireIn)
		utils.Check(err)
		fmt.Printf("File '%s' pushed to '%s' for current job.\n", src, dst)
	},
}

var PushWorkflowCmd = &cobra.Command{
	Use:   "workflow [SOURCE PATH]",
	Short: "Upload a workflow file to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		src := args[0]

		dst, err := cmd.Flags().GetString("destination")
		utils.Check(err)

		expireIn, err := cmd.Flags().GetString("expire-in")
		utils.Check(err)

		dst, src, err = pushFileGCS(utils.WORKFLOW, dst, src, expireIn)
		utils.Check(err)
		fmt.Printf("File '%s' pushed to '%s' for current workflow.\n", src, dst)
	},
}

var PushProjectCmd = &cobra.Command{
	Use:   "project [SOURCE PATH]",
	Short: "Upload a project file to the storage.",
	Long:  ``,
	Args:  cobra.ExactArgs(1),

	Run: func(cmd *cobra.Command, args []string) {
		src := args[0]

		dst, err := cmd.Flags().GetString("destination")
		utils.Check(err)

		expireIn, err := cmd.Flags().GetString("expire-in")
		utils.Check(err)

		dst, src, err = pushFileGCS(utils.PROJECT, dst, src, expireIn)
		utils.Check(err)
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
}
