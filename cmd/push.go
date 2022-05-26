package cmd

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	errutil "github.com/semaphoreci/artifact/pkg/errors"
	"github.com/semaphoreci/artifact/pkg/files"
	"github.com/semaphoreci/artifact/pkg/hub"
	"github.com/semaphoreci/artifact/pkg/storage"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const ExpireInDescription = `removes the files after the given amount of time.

- Nd for N days
- Nw for N weeks
- Nm for N months
- Ny for N years

WARNING: This is an obsolete flag and has no effect.
Set up a retention policy in your project instead.
Docs: https://docs.semaphoreci.com/essentials/artifacts/#artifact-retention-policies.
`

// pushCmd represents the push command
var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Stores a file or directory in the storage for later use",
	Long: `You may store project, workflow or job related files, that you can use
while the rest of the semaphore process, or after it.`,
}

func runPushForCategory(cmd *cobra.Command, args []string, resolver *files.PathResolver) (*files.ResolvedPath, error) {
	hubClient, err := hub.NewClient()
	errutil.Check(err)

	localSource, err := getSrc(cmd, args)
	errutil.Check(err)

	destinationOverride, err := cmd.Flags().GetString("destination")
	errutil.Check(err)

	force, err := cmd.Flags().GetBool("force")
	errutil.Check(err)

	expireIn, err := cmd.Flags().GetString("expire-in")
	errutil.Check(err)
	if len(expireIn) != 0 {
		displayWarningThatExpireInIsNoLongerSupported()
	}

	return storage.Push(hubClient, resolver, storage.PushOptions{
		SourcePath:          localSource,
		DestinationOverride: destinationOverride,
		Force:               force,
	})
}

func displayWarningThatExpireInIsNoLongerSupported() {
	fmt.Println("")
	fmt.Println("WARNING: The --expire-in flag is obsolete and will have no efffect.")
	fmt.Println("Use artifact retention policies to control the lifetime of artifacts.")
	fmt.Println("Docs: https://docs.semaphoreci.com/essentials/artifacts/#artifact-retention-policies")
	fmt.Println("")
}

func NewPushJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job [SOURCE PATH]",
		Short: "Uploads a job file or directory to the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			jobId, err := cmd.Flags().GetString("job-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeJob, jobId)
			errutil.Check(err)

			paths, err := runPushForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error pushing artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pushed artifact for current job.\n")
			log.Infof("* Local source: %s.\n", paths.Source)
			log.Infof("* Remote destination: %s.\n", paths.Destination)
		},
	}

	cmd.Flags().StringP("destination", "d", "", "rename the file while uploading")
	cmd.Flags().BoolP("force", "f", false, "force overwrite")
	cmd.Flags().StringP("expire-in", "e", "", ExpireInDescription)
	cmd.Flags().StringP("job-id", "j", "", "set explicit job id")

	return cmd
}

func NewPushWorkflowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflow [SOURCE PATH]",
		Short: "Uploads a workflow or directory file to the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			workflowId, err := cmd.Flags().GetString("workflow-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeWorkflow, workflowId)
			errutil.Check(err)

			paths, err := runPushForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error pushing artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pushed artifact for current job.\n")
			log.Infof("* Local source: %s.\n", paths.Source)
			log.Infof("* Remote destination: %s.\n", paths.Destination)
		},
	}

	cmd.Flags().StringP("destination", "d", "", "rename the file while uploading")
	cmd.Flags().BoolP("force", "f", false, "force overwrite")
	cmd.Flags().StringP("expire-in", "e", "", ExpireInDescription)
	cmd.Flags().StringP("workflow-id", "w", "", "set explicit workflow id")

	return cmd
}

func NewPushProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project [SOURCE PATH]",
		Short: "Upload a project file or directory to the storage.",
		Long:  ``,
		Args:  cobra.ExactArgs(1),

		Run: func(cmd *cobra.Command, args []string) {
			projectId, err := cmd.Flags().GetString("project-id")
			errutil.Check(err)

			resolver, err := files.NewPathResolver(files.ResourceTypeProject, projectId)
			errutil.Check(err)

			paths, err := runPushForCategory(cmd, args, resolver)
			if err != nil {
				log.Errorf("Error pushing artifact: %v\n", err)
				errutil.Exit(1)
				return
			}

			log.Info("Successfully pushed artifact for current job.\n")
			log.Infof("* Local source: %s.\n", paths.Source)
			log.Infof("* Remote destination: %s.\n", paths.Destination)
		},
	}

	cmd.Flags().StringP("destination", "d", "", "rename the file while uploading")
	cmd.Flags().BoolP("force", "f", false, "force overwrite")
	cmd.Flags().StringP("expire-in", "e", "", ExpireInDescription)
	cmd.Flags().StringP("project-id", "p", "", "set explicit project id")

	return cmd
}

func init() {
	rootCmd.AddCommand(pushCmd)
	pushCmd.AddCommand(NewPushJobCmd())
	pushCmd.AddCommand(NewPushWorkflowCmd())
	pushCmd.AddCommand(NewPushProjectCmd())
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
