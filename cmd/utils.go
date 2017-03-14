package cmd

import (
	"fmt"
	"os"

	"github.com/bypasslane/gzr/comms"
	"github.com/spf13/cobra"
)

// Flag var for holding namespace info
var namespace string

// webPort is the port that the web interface will run on
var webPort int

// imageStore is the backing for image data storage
var imageStore comms.GzrMetadataStore

// available interfaces for image storage
var registeredInterfaces = make(map[string]func() (comms.GzrMetadataStore, error))

// imageManager is the backing for image managing (building, pushing)
var imageManager comms.ImageManager

// er prints an error message and exits. Lifted from Cobra source.
func er(msg interface{}) {
	fmt.Println("Error:", msg)
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
