package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/bypasslane/gzr/comms"
	"github.com/ericchiang/k8s/apis/extensions/v1beta1"
)

func emptyDeploymentsList() (**v1beta1.DeploymentList, error) {
	return nil, nil
}

func populatedDeploymentsList() (*v1beta1.DeploymentList, error) {
	return &v1beta1.DeploymentList{}, nil
}

func emptyGetDeployment(deploymentName string) (*v1beta1.Deployment, error) {
	return nil, comms.ErrDeploymentNotFound
}

func populatedGetDeployment(deploymentName string) (*v1beta1.Deployment, error) {
	return &v1beta1.Deployment{}, nil
}

func successfulUpdateDeployment(newDeployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return &v1beta1.Deployment{}, nil
}

func failUpdateDeploymentNoDeployment(newDeployment *v1beta1.Deployment) (*v1beta1.Deployment, error) {
	return &v1beta1.Deployment{}, comms.ErrDeploymentNotFound
}

// Sends an HTTP request to provided server:
// GET /deployments
func getDeploymentsList(server *httptest.Server) (*http.Response, error) {
	client := new(http.Client)
	req, _ := http.NewRequest("GET", server.URL+"/deployments", nil)
	return client.Do(req)
}

// Sends an HTTP request to provided server:
// GET /deployments/{name}
func getDeployment(server *httptest.Server) (*http.Response, error) {
	client := new(http.Client)
	req, _ := http.NewRequest("GET", server.URL+"/deployments/name", nil)
	return client.Do(req)
}

// Sends an HTTP request to provided server:
// PUT /deployments/{name}
func updateDeployment(server *httptest.Server) (*http.Response, error) {
	client := new(http.Client)
	payloadSource := `{"container_name": "foobaricus", "image": "foobar:1.2.3"}`
	reader := strings.NewReader(payloadSource)
	req, _ := http.NewRequest("PUT", server.URL+"/deployments/name", reader)
	return client.Do(req)
}
