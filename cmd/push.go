package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/semaphoreci/artifact/pkg/gcs"
	errutil "github.com/semaphoreci/artifact/pkg/util/err"
	"github.com/semaphoreci/artifact/pkg/util/log"
	pathutil "github.com/semaphoreci/artifact/pkg/util/path"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Stores a file or directory in the storage for later use",
	Long: `You may store project, workflow or job related files, that you can use
while the rest of the semaphore process, or after it.`,
}

func runPushForCategory(cmd *cobra.Command, args []string, category, catID string) (string, string) {
	err := pathutil.InitPathID(category, catID)
	errutil.Check(err)

	src, err := getSrc(cmd, args)
	errutil.Check(err)

	dst, err := cmd.Flags().GetString("destination")
	errutil.Check(err)

	force, err := cmd.Flags().GetBool("force")
	errutil.Check(err)

	expireIn, err := cmd.Flags().GetString("expire-in")
	errutil.Check(err)
	if len(expireIn) != 0 {
		displayWarningThatExpireInIsNoLongerSupported()
	}

	dst, src = gcs.PushPaths(filepath.ToSlash(dst), filepath.ToSlash(src))
	if ok := gcs.PushGCS(dst, src, force); !ok {
		os.Exit(1) // error already logged
	}
	return dst, src
}

func displayWarningThatExpireInIsNoLongerSupported() {
	fmt.Println("")
	fmt.Println("WARNING: The --expire-in flag is obsolete and will have no efffect.")
	fmt.Println("Use artifact retention policies to control the lifetime of artifacts.")
	fmt.Println("Docs: https://docs.semaphoreci.com/essentials/artifacts/#artifact-retention-policies")
	fmt.Println("")
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
		dst, src := runPushForCategory(cmd, args, pathutil.JOB, catID)
		log.Info("successful push for current job", zap.String("source", src), zap.String("destination", dst))
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
		dst, src := runPushForCategory(cmd, args, pathutil.WORKFLOW, catID)
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
		dst, src := runPushForCategory(cmd, args, pathutil.PROJECT, catID)
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

	desc = `removes the files after the given amount of time.

  - Nd for N days
  - Nw for N weeks
  - Nm for N months
  - Ny for N years

WARNING: This is an obsolete flag and has no effect.
Set up a retention policy in your project instead.
Docs: https://docs.semaphoreci.com/essentials/artifacts/#artifact-retention-policies.
`
	PushJobCmd.Flags().StringP("expire-in", "e", "", desc)
	PushWorkflowCmd.Flags().StringP("expire-in", "e", "", desc)
	PushProjectCmd.Flags().StringP("expire-in", "e", "", desc)

	PushJobCmd.Flags().StringP("job-id", "j", "", "set explicit job id")
	PushWorkflowCmd.Flags().StringP("workflow-id", "w", "", "set explicit workflow id")
	PushProjectCmd.Flags().StringP("project-id", "p", "", "set explicit project id")
}

func getSrc(cmd *cobra.Command, args []string) (string, error) {
	if shouldUseStdin() {
		log.Debug("Detected stdin, saving it to a temporary file...")
		return saveStdinToTempFile()
	}

	return args[0], nil
}

func shouldUseStdin() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	return (stat.Mode() & os.ModeCharDevice) == 0
}

func saveStdinToTempFile() (string, error) {
	tmpFile, err := ioutil.TempFile("", "*")
	if err != nil {
		log.Error("Error creating temporary file to read stdin", zap.Error(err))
		return "", err
	}

	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 0, 8)

	for {
		nRead, err := r.Read(buf[:cap(buf)])
		buf = buf[:nRead]

		// nothing was read and no error was thrown, so just try again
		if nRead == 0 && err == nil {
			continue
		}

		// there's nothing more to read
		if err == io.EOF {
			break
		}

		// Something went wrong when reading
		if err != nil {
			log.Error("Error reading stdin", zap.Error(err))
			return "", err
		}

		_, err = tmpFile.Write(buf)
		if err != nil {
			log.Error("Error writing to temp file", zap.Error(err))
			return "", err
		}
	}

	return tmpFile.Name(), nil
}
