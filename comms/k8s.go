package comms

import (
	"context"
	"encoding/json"
	e "errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"

	"github.com/ericchiang/k8s"
	"github.com/ericchiang/k8s/apis/extensions/v1beta1"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

var (
	ErrDeploymentNotFound       = e.New("Requested deployment couldn't be found")
	ErrNoDeploymentsInNamespace = e.New("No deployments found in specified namespace")
)

// K8sClient is a wrapper around the *k8s.Client
type K8sClient struct {
	K *k8s.Client
}

// K8sCommunicator defines an interface for retrieving data from a k8s cluster
type K8sCommunicator interface {
	// ListDeployments returns the list of Deployments in the cluster
	ListDeployments() (*v1beta1.DeploymentList, error)
	// GetDeployment returns the Deployment matching the given name
	GetDeployment(string) (*v1beta1.Deployment, error)
	// UpdateDeployment updates the Deployment's container in the manner specified by the argument
	UpdateDeployment(*v1beta1.Deployment) (*v1beta1.Deployment, error)
}

// NewK8sClient returns an in-cluster K8sClient, if gzr is ran from outside of a K8's cluster it
// falls back to looking at users kube config.
func NewK8sClient() (*K8sClient, error) {
	client := new(K8sClient)
	k, err := k8s.NewInClusterClient()

	// https://github.com/ericchiang/k8s/blob/master/client.go#L251
	// Currently there aren't any error types to check against, without warning the user we try to
	// create a client using kube config when in-cluster client creation fails
	if err != nil {
		kubeconfigPath := fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
		data, err := ioutil.ReadFile(kubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("read kubeconfig: %v", err)
		}

		// Unmarshal YAML into a Kubernetes config object.
		var config k8s.Config
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("unmarshal kubeconfig: %v", err)
		}
		k, err = k8s.NewClient(&config)
		if err != nil {
			panic(err.Error())
		}
	}
	client.K = k
	return client, nil
}

// GetDeployment returns a v1beta1.Deployment matching the deploymentName in current namespace
func (client *K8sClient) GetDeployment(deploymentName string) (*v1beta1.Deployment, error) {
	deployment, err := client.K.ExtensionsV1Beta1().GetDeployment(context.Background(), deploymentName, client.K.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get deployment %q in namespace %q", deploymentName, client.K.Namespace)
	}
	return deployment, err
}

// ListDeployments returns a list of Deployments in current namespace
func (client *K8sClient) ListDeployments() (*v1beta1.DeploymentList, error) {
	deploymentList, err := client.K.ExtensionsV1Beta1().ListDeployments(context.Background(), client.K.Namespace)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to list deployments in namespace %q", client.K.Namespace)
	}
	return deploymentList, err
}

// UpdateDeployment updates a Deployment on the server to the structure represented by the argument
// TODO: verify that requested image exists in the store
// TODO: verify that requested image exists in the registry
func (client *K8sClient) UpdateDeployment(newDeployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	deployment, err := client.K.ExtensionsV1Beta1().GetDeployment(context.Background(), *newDeployment.Metadata.Name, client.K.Namespace)
	// no Name in ObjectMeta means it was returned empty
	if *deployment.Metadata.Name == "" {
		return deployment, errors.WithStack(ErrDeploymentNotFound)
	}

	deployment, err = client.K.ExtensionsV1Beta1().UpdateDeployment(context.Background(), newDeployment)

	if err != nil {
		return deployment, errors.Wrap(err, "Failed to update deployment")
	}

	return deployment, nil
}

// SerializeDeployForCLI takes an io.Writer and writes templatized data to it representing a Deployment
func SerializeDeployForCLI(deploy *v1beta1.Deployment, wr io.Writer) error {
	t := template.New("Deployment CLI")
	t, _ = t.Parse(`-------------------------
Deployment: {{.Metadata.Name}}
  - replicas: {{.Spec.Replicas}}
  - containers: {{range .Spec.Template.Spec.Containers}}
    --name:  {{.Name}}
    --image: {{.Image}}
{{end}}
`)
	return errors.Wrap(t.Execute(wr, deploy), "Failed to serialize deployment ")
}

// SerializeForWire returns a JSON representation of the obj
func SerializeForWire(obj interface{}) ([]byte, error) {
	data, err := json.Marshal(obj)
	return data, errors.Wrap(err, "Failed to convert object to json")
}
