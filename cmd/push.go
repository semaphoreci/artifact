package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	errutil "github.com/semaphoreci/artifact/pkg/err"
	"github.com/semaphoreci/artifact/pkg/hub"
	pathutil "github.com/semaphoreci/artifact/pkg/path"
	"github.com/semaphoreci/artifact/pkg/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Stores a file or directory in the storage for later use",
	Long: `You may store project, workflow or job related files, that you can use
while the rest of the semaphore process, or after it.`,
}

func runPushForCategory(cmd *cobra.Command, args []string, category, catID string) (string, string) {
	hubClient, err := hub.NewClient()
	errutil.Check(err)

	err = pathutil.InitPathID(category, catID)
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

	dst, src = storage.PushPaths(filepath.ToSlash(dst), filepath.ToSlash(src))
	if ok := storage.Push(hubClient, dst, src, force); !ok {
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
		log.Infof("Successfully pushed '%s' as '%s' for current job.\n", src, dst)
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
		log.Infof("Successfully pushed '%s' as '%s' for current workflow.\n", src, dst)
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
		log.Infof("Successfully pushed '%s' as '%s' for current project.\n", src, dst)
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
		log.Errorf("Error creating temporary file to read stdin: %v\n", err)
		return "", err
	}

	r := bufio.NewReader(os.Stdin)
	buf := make([]byte, 0, 4*1024)

	for {
		nRead, err := r.Read(buf[:cap(buf)])
		buf = buf[:nRead]

		if nRead == 0 {
			// nothing was read and no error was thrown, so just try again
			if err == nil {
				continue
			}

			// there's nothing more to read
			if err == io.EOF {
				break
			}

			// nothing was read and we had an error
			log.Errorf("Error reading stdin: %v\n", err)
			return "", err
		}

		// something was read, but we still got an error
		if err != nil && err != io.EOF {
			log.Errorf("Error reading stdin: %v\n", err)
			return "", err
		}

		// no errors when reading from stdin, just write it to the temporary file
		_, err = tmpFile.Write(buf)
		if err != nil {
			log.Errorf("Error writing to temp file: %v\n", err)
			return "", err
		}
	}

	return tmpFile.Name(), nil
}
