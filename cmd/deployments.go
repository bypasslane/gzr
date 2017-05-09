package cmd

import (
	"fmt"
	"os"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/bypasslane/gzr/comms"
	"github.com/spf13/cobra"
)

// Package-global k8s connection
var k8sConn *comms.K8sConnection

// originalContext to keep track of original context
var originalContext string

// changedContext keeps track of if the context was changed
var changedContext bool

// deploymentsCmd represents the deployments command
var deploymentsCmd = &cobra.Command{
	Use:   "deployments [subcommand]",
	Short: "Manage k8s Deployments",
	Long: `Used to get information on Deployments or update them

deployments list
deployments get <DEPLOYMENT NAME>
deployments update <DEPLOYMENT_NAME> <CONTAINER_NAME> <IMAGE>
	`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		changedContext = false
		cli, err := clientcmd.LoadFromFile(clientcmd.RecommendedHomeFile)
		if err != nil {
			er("Cannot load ~/.kube/config")
		}
		currentContext := cli.CurrentContext
		originalContext = currentContext
		// If a context was set from a flag and it's different than currentContext,
		// swap them for the duration of the command
		if currentContext != context && context != "" {
			changeContext(context)
			changedContext = true
		}
		setNamespace(cli)
		var connErr error
		k8sConn, connErr = comms.NewK8sConnection(namespace)
		if connErr != nil {
			// TODO: figure out the Cobra way to handle this
			erWithDetails(connErr, "problem establishing k8s connection")
		}
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if changedContext {
			changeContext(originalContext)
		}
	},
}

// deploymentsListCmd returns a list of deployments
var deploymentsListCmd = &cobra.Command{
	Use:   "list [flags]",
	Short: "List k8s Deployments",
	Long: `Used to get ReplicaSet and PodSpec information on all Deployments.

deployments list
	`,
	Run: func(cmd *cobra.Command, args []string) {
		listDeploymentsHandler()
	},
}

// deploymentGetCmd returns a list of deployments
var deploymentGetCmd = &cobra.Command{
	Use:   "get <DEPLOYMENT_NAME> [flags]",
	Short: "Get a k8s Deployment by name",
	Long: `Used to get a single Deployment by name, showing ReplicaSet
and PodSpec information.

deployments get mah-deployment
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 1 {
			erBadUsage("Not enough arguments", cmd)
		}
		getDeploymentHandler(args[0])
	},
}

// deploymentUpdateCmd returns a list of deployments
var deploymentUpdateCmd = &cobra.Command{
	Use:   "update <DEPLOYMENT_NAME> <CONTAINER_NAME> <IMAGE> [flags]",
	Short: "Update a container in a Deployment to a specific image",
	Long: `Used to update a particular container in the Deployment's PodSpec by name.

deployments update mah-deployment some-pod-container coolthing:latest
	`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) < 3 {
			erBadUsage("Not enough arguments", cmd)
		}
		updateDeploymentHandler(namespace, args[0], args[1], args[2])
	},
}

// updateDeploymentHandler updates a Deployment container with the info described by the DeploymentContainerInfo argument
func updateDeploymentHandler(namespace string, deploymentName string, containerName string, image string) {
	dci := &comms.DeploymentContainerInfo{
		Namespace:      namespace,
		DeploymentName: deploymentName,
		ContainerName:  containerName,
		Image:          image,
	}
	deployment, err := k8sConn.UpdateDeployment(dci)

	if err != nil {
		erWithDetails(err, fmt.Sprintf("There was a problem updating container %q on deployment %q", containerName, deploymentName))
	}
	deployment.SerializeForCLI(os.Stdout)
}

// getDeploymentHandler fetches
func getDeploymentHandler(deploymentName string) {
	deployment, err := k8sConn.GetDeployment(deploymentName)
	if err != nil {
		erWithDetails(err, fmt.Sprintf("There was a problem retrieving deployment %q", deploymentName))
	}
	deployment.SerializeForCLI(os.Stdout)
}

// listDeploymentsHandler fetches Deployments and prints them to the CLI
func listDeploymentsHandler() {
	dlist, err := k8sConn.ListDeployments()
	if err != nil {
		erWithDetails(err, "Error retrieving deployments")
	}
	for _, deployment := range dlist.Deployments {
		deployment.SerializeForCLI(os.Stdout)
	}
}

func init() {
	deploymentsCmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "", "namespace to look for Deployments in")
	deploymentsCmd.PersistentFlags().StringVarP(&context, "context", "c", "", "context to look for Deployments in")
	deploymentsCmd.AddCommand(deploymentsListCmd)
	deploymentsCmd.AddCommand(deploymentGetCmd)
	deploymentsCmd.AddCommand(deploymentUpdateCmd)
	RootCmd.AddCommand(deploymentsCmd)
}
