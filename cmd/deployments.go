package cmd

import (
	"fmt"
	"os"

	"github.com/bypasslane/gzr/comms"
	"github.com/spf13/cobra"
)

// Package-global k8s connection
var k8sClient *comms.K8sClient

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
		var err error
		k8sClient, err = comms.NewK8sClient()
		if err != nil {
			// TODO: figure out the Cobra way to handle this
			erWithDetails(err, "problem establishing k8s connection")
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
		updateDeploymentHandler(args[0], args[1], args[2])
	},
}

// updateDeploymentHandler updates a Deployment container with the info described by the DeploymentContainerInfo argument
func updateDeploymentHandler(deploymentName string, containerName string, image string) {
	var foundContainer bool

	deployment, err := k8sClient.GetDeployment(deploymentName)
	if err != nil {
		erWithDetails(err, fmt.Sprintf("There was a problem retrieving deployment %q", deploymentName))
	}

	for containerIndex, container := range deployment.Spec.Template.Spec.Containers {
		if *container.Name == containerName {
			foundContainer = true
			*deployment.Spec.Template.Spec.Containers[containerIndex].Image = image
			break
		}
	}
	if !foundContainer {
		erWithDetails(err, fmt.Sprintf("Could not find container with name %q", containerName))
	}

	deployment, err = k8sClient.UpdateDeployment(deployment)

	if err != nil {
		erWithDetails(err, fmt.Sprintf("There was a problem updating container %q on deployment %q", containerName, deploymentName))
	}

	comms.SerializeDeployForCLI(deployment, os.Stdout)
}

// getDeploymentHandler fetches
func getDeploymentHandler(deploymentName string) {
	deployment, err := k8sClient.GetDeployment(deploymentName)
	if err != nil {
		erWithDetails(err, fmt.Sprintf("There was a problem retrieving deployment %q", deploymentName))
	}
	comms.SerializeDeployForCLI(deployment, os.Stdout)
}

// listDeploymentsHandler fetches Deployments and prints them to the CLI
func listDeploymentsHandler() {
	dlist, err := k8sClient.ListDeployments()
	if err != nil {
		erWithDetails(err, "Error retrieving deployments")
	}
	for _, deployment := range dlist.GetItems() {
		comms.SerializeDeployForCLI(deployment, os.Stdout)
	}
}

func init() {
	deploymentsCmd.AddCommand(deploymentsListCmd)
	deploymentsCmd.AddCommand(deploymentGetCmd)
	deploymentsCmd.AddCommand(deploymentUpdateCmd)
	RootCmd.AddCommand(deploymentsCmd)
}
