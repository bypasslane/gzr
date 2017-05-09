package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"k8s.io/client-go/tools/clientcmd/api"

	log "github.com/Sirupsen/logrus"
	"github.com/bypasslane/gzr/comms"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Flag var for holding namespace info
var namespace string

// Flag var for holding context info
var context string

// webPort is the port that the web interface will run on
var webPort int

var logLevel string
var logFormat string

// imageStore is the backing for image data storage
var imageStore comms.GzrMetadataStore

// available interfaces for image storage
var registeredInterfaces = make(map[string]func() (comms.GzrMetadataStore, error))

// imageManager is the backing for image managing (building, pushing)
var imageManager comms.ImageManager

// er prints an error message and exits. Lifted from Cobra source.
func er(msg interface{}) {
	log.Error(msg)
	os.Exit(-1)
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

// erWithDetails prints the error message and the stacktrace if it has one then exits. Should be restricted to CLI usage
func erWithDetails(err error, msg interface{}) {
	logEntry := log.WithField("error", err)
	if err, ok := err.(stackTracer); ok {
		logEntry = logEntry.WithField("stacktrace", fmt.Sprintf("%+v", err.(stackTracer).StackTrace()))
	}
	logEntry.Error(msg)
	os.Exit(-1)
}

// erBadUsage prints a message and the usage for the command
func erBadUsage(msg string, cmd *cobra.Command) {
	fmt.Println(msg)
	fmt.Println(cmd.Use)
	os.Exit(1)
}

// notify sends a formatted information line to stdout
func notify(msg string) {
	fmt.Printf("[-] %s\n", msg)
}

func setupImageStore() {
	storeType := viper.GetString("datastore.type")
	if storeType == "" {
		er("Must supply a datastore type in config file")
	}

	if viper.GetString("repository") == "" {
		er("Must provide \"repository\" setting in config file")
	}

	var storeCreator func() (comms.GzrMetadataStore, error)
	if creator, ok := registeredInterfaces[storeType]; !ok {
		er(fmt.Sprintf("%s is not a valid datastore type", storeType))
	} else {
		storeCreator = creator
	}
	newStore, err := storeCreator()
	if err != nil {
		erWithDetails(err, "Failed to initialize store")
	}
	imageStore = newStore
}

// setNamespace checks the current context in k8's config if one has not been passed into
// the command
func setNamespace(cli *api.Config) {
	if namespace == "" { // If it's set from the flag, don't do anything
		currentNamespace := cli.Contexts[cli.CurrentContext].Namespace
		if currentNamespace != "" { // If it's set in the context, set it
			namespace = currentNamespace
		} else {
			namespace = "default"
		}
	}
}

// changeContext actually calls `kubectl config use-context <context>` in order
// to change the context of the local kubernetes config
func changeContext(context string) {
	ctxCmd := exec.Command("kubectl", "config", "use-context", context)
	err := ctxCmd.Run()
	if err != nil {
		er(err)
	}
}
