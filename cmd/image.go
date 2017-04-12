// Image is the main command for handing gzr relevant data to inform it about
// image artifacts that cannot be discovered automatically

package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/bypasslane/gzr/comms"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var latest bool

var imageCmd = &cobra.Command{
	Use:   "image (store|get|delete)",
	Short: "manage information about images",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		setupImageStore()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		imageStore.Cleanup()
	},
}

var storeCmd = &cobra.Command{
	Use:   "store IMAGE_NAME:VERSION METADATA_PATH",
	Short: "Store metadata about an image for gzr to track",
	Long: `Used to store metadata about an image for gzr to track. The name must be formatted as NAME:VERSION.
Repeated store calls with the same VERSION on the same day will only keep the newest, but different days will be stored.
In short, only one version per day is allowed.

The structure of the JSON at the METADATA_PATH should be as follows:
{
    "git-commit": <string>,
    "git-tag": [<string>, <string>, ...],
    "git-annotation": [<string>, <string>, ...],
    "git-origin": <string>,
    "created-at": <ISO 8601-formatted string>
}`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 2 {
			erBadUsage("Must provide IMAGE_NAME:VERSION and METADATA_PATH", cmd)
		}
		reader, err := os.Open(args[1])
		if err != nil {
			erWithDetails(err, "Could not read metadata file")
		}
		meta, err := comms.CreateMeta(reader)
		if err != nil {
			erWithDetails(err, "Create meta failed")
		}
		err = imageStore.StartTransaction()
		if err != nil {
			erWithDetails(err, "failed to start transaction")
		}
		err = imageStore.Store(args[0], meta)
		if err != nil {
			erWithDetails(err, "Error storing image")
		}
		err = imageStore.CommitTransaction()
		if err != nil {
			erWithDetails(err, "failed to commit transaction")
		}
		fmt.Printf("Stored %s\n", args[0])
	},
}

var getCmd = &cobra.Command{
	Use:   "get IMAGE_NAME",
	Short: "Get data about the stored images under a particular name",
	Long: `Get all metadata about the stored images under a particular name,
including all versions held within gzr`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			erBadUsage("Must provide IMAGE_NAME", cmd)
		}
		name := fmt.Sprintf("%s/%s", viper.GetString("repository"), args[0])
		if latest {
			image, err := imageStore.GetLatest(name)
			if err != nil {
				erWithDetails(err, "Failed to get latest image")
			}
			image.SerializeForCLI(os.Stdout)
		} else {
			images, err := imageStore.List(name)
			if err != nil {
				erWithDetails(err, "Failed to get images")
			}
			images.SerializeForCLI(os.Stdout)
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete IMAGE_NAME:VERSION",
	Short: "Delete metadata about an image:version within gzr",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			erBadUsage("Must provide IMAGE_NAME:VERSION", cmd)
		}
		splitName := strings.Split(args[0], ":")
		if len(splitName) != 2 {
			er("IMAGE_NAME must be formatted as NAME:VERSION and must contain only the seperating colon")
		}
		err := imageStore.StartTransaction()
		if err != nil {
			erWithDetails(err, "Failed to start transaction")
		}
		name := fmt.Sprintf("%s/%s", viper.GetString("repository"), args[0])
		deleted, err := imageStore.Delete(name)
		if err != nil {
			erWithDetails(err, "Failed to delete image")
		}
		err = imageStore.CommitTransaction()
		if err != nil {
			erWithDetails(err, "Failed to commit transaction")
		}
		fmt.Printf("Deleted %d\n", deleted)
	},
}

func init() {
	getCmd.Flags().BoolVarP(&latest, "latest", "l", false, "option to just get the latest image")
	imageCmd.AddCommand(storeCmd)
	imageCmd.AddCommand(getCmd)
	imageCmd.AddCommand(deleteCmd)
	RootCmd.AddCommand(imageCmd)
}
