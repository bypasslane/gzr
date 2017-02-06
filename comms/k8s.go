package comms

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"text/template"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	ErrContainerNotFound = errors.New("requested container couldn't be found")
)

// GzrDeployment is just here to let us declare methods on k8s Deployments
type GzrDeployment v1beta1.Deployment

// GzrDeploymentList is a collection of GzrDeployments
type GzrDeploymentList struct {
	Deployments []GzrDeployment `json:"deployments"`
}

// Serializer knows how to serialize for web (JSON) and CLI (templatized strings)
type Serializer interface {
	// SerializeForCLI writes templatized information to the provided io.Writer
	SerializeForCLI(io.Writer) error
	// SerializeForWire kicks out JSON as a byte slice
	SerializeForWire() ([]byte, error)
}

// DeploymentContainerInfo holds information about a Deployment sufficient for updating a Pod's container by name
type DeploymentContainerInfo struct {
	// Namespace is the Deployment's k8s namespace
	Namespace string
	// DeploymentName is the name of the Deployment
	DeploymentName string
	// ContainerName is the name of a Pod's container in the Deployment spec
	ContainerName string
	// Image is the name of the image (current or intended) for the container identified by ContainerName
	Image string
}

// K8sCommunicator defines an interface for retrieving data from a k8s cluster
type K8sCommunicator interface {
	// ListDeployments returns the list of Deployments in the cluster
	ListDeployments(string) ([]GzrDeployment, error)
	// GetDeployment returns the Deployment matching the given name
	GetDeployment(string, string) (GzrDeployment, error)
}

// K8sConnection implements the K8sCommunicator interface and holds a live connection to a k8s cluster
type K8sConnection struct {
	// clientset is a collection of Kubernetes API clients
	clientset *kubernetes.Clientset
	// Namespace is the k8s namespace active for this connection used to talk
	Namespace string
}

// NewK8sConnection returns a K8sConnection with an active v1.Clientset.
//   - assumes that $HOME/.kube/config contains a legit Kubernetes config for an healthy k8s cluster.
//   - panics if the configuration can't be used to connect to a k8s cluster.
func NewK8sConnection(namespace string) (*K8sConnection, error) {
	var k *K8sConnection
	kubeconfig := fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		return k, err
	}

	k = &K8sConnection{
		clientset: clientset,
		Namespace: namespace,
	}

	return k, nil
}

// GetDeployment returns a GzrDeployment matching the deploymentName in the given namespace
func (k *K8sConnection) GetDeployment(namespace string, deploymentName string) (*GzrDeployment, error) {
	var gd *GzrDeployment
	deployment, err := k.clientset.ExtensionsV1beta1().Deployments(namespace).Get(deploymentName)
	if err != nil {
		return gd, err
	}
	gdp := GzrDeployment(*deployment)
	gd = &gdp

	return gd, err
}

// UpdateDeployment updates a Deployment on the server to the structure represented by the argument
// TODO: verify that requested image exists in the store
// TODO: verify that requested image exists in the registry
func (k *K8sConnection) UpdateDeployment(dci *DeploymentContainerInfo) (*GzrDeployment, error) {
	var gd *GzrDeployment
	var containerIndex int
	found := false

	deployment, err := k.clientset.ExtensionsV1beta1().Deployments(dci.Namespace).Get(dci.DeploymentName)
	for index, container := range deployment.Spec.Template.Spec.Containers {
		if container.Name == dci.ContainerName {
			containerIndex = index
			found = true
		}
	}

	if !found {
		return gd, ErrContainerNotFound
	}

	deployment.Spec.Template.Spec.Containers[containerIndex].Image = dci.Image
	deployment, err = k.clientset.ExtensionsV1beta1().Deployments(dci.Namespace).Update(deployment)

	if err != nil {
		return gd, err
	}

	gdp := GzrDeployment(*deployment)
	gd = &gdp

	return gd, err
}

// ListDeployments returns the active k8s Deployments for the given namespace
func (k *K8sConnection) ListDeployments(namespace string) (*GzrDeploymentList, error) {
	var gzrDeploymentList GzrDeploymentList
	deploymentList, err := k.clientset.ExtensionsV1beta1().Deployments(namespace).List(v1.ListOptions{})
	if err != nil {
		return &gzrDeploymentList, err
	}

	for _, deployment := range deploymentList.Items {
		gzrDeploymentList.Deployments = append(gzrDeploymentList.Deployments, GzrDeployment(deployment))
	}

	return &gzrDeploymentList, nil
}

// SerializeForCLI takes an io.Writer and writes templatized data to it representing a Deployment
func (d GzrDeployment) SerializeForCLI(wr io.Writer) error {
	return d.cliTemplate().Execute(wr, d)
}

// cliTemplate returns the template that will be used for serializing Deployment data for display in the CLI
func (d GzrDeployment) cliTemplate() *template.Template {
	t := template.New("Deployment CLI")
	t, _ = t.Parse(`-------------------------
Deployment: {{.ObjectMeta.Name}}
  - replicas: {{.Spec.Replicas}}
  - containers: {{range .Spec.Template.Spec.Containers}}
    --name:  {{.Name}}
    --image: {{.Image}}
{{end}}
`)
	return t
}

// SerializeForWire returns a JSON representation of the Deployment
func (d GzrDeployment) SerializeForWire() ([]byte, error) {
	return json.Marshal(d)
}

// SerializeForWire returns a JSON representation of the DeploymentList
func (dl *GzrDeploymentList) SerializeForWire() ([]byte, error) {
	return json.Marshal(dl)
}
