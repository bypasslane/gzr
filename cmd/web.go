package cmd

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/bypasslane/gzr/comms"
	"github.com/bypasslane/gzr/controllers"
	"github.com/spf13/cobra"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:   "web",
	Short: "Stand up the gzr web interface",
	Long: `Use gzr functionality from inside the browser
gzr web
gzr web --port=<CUSTOM_PORT_NUMBER>
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		log.SetFormatter(&log.JSONFormatter{})
		var connErr error
		k8sConn, connErr = comms.NewK8sConnection(namespace)
		if connErr != nil {
			// TODO: figure out the Cobra way to handle this
			erWithDetails(connErr, "Problem establishing k8s connection")
		}
		setupImageStore()
	},
	Run: func(cmd *cobra.Command, args []string) {
		bindAndRun()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		imageStore.Cleanup()
	},
}

// bindAndRun starts the server
func bindAndRun() {
	portString := fmt.Sprintf(":%v", webPort)
	fmt.Printf("[-] Listening on %v\n", portString)
	http.ListenAndServe(portString, controllers.App(k8sConn, imageStore))
}

func init() {
	RootCmd.AddCommand(webCmd)
	webCmd.Flags().IntVarP(&webPort, "port", "p", 9393, "the port to run the Gozer web interface on")
	webCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "namespace to look for Deployments in")
}
