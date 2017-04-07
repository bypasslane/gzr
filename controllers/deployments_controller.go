package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/bypasslane/gzr/comms"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

// UpdateDeploymentUserType represents the payload of data that will come in from
// the client for updating a specific container in a Deployment spec
type UpdateDeploymentUserType struct {
	ContainerName string `json:"container_name"`
	Image         string `json:"image"`
}

// listDeploymentsHandler lists deployments in the Kubernetes instance
func listDeploymentsHandler(k8sConn comms.K8sCommunicator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deployments, err := k8sConn.ListDeployments()
		// TODO: differentiate between legit errors and unhandleable errors
		if err != nil {
			errors.Wrap(err, "")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if deployments == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jsonData, err := deployments.SerializeForWire()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(jsonData)
	})
}

// getDeploymentHandler gets a single Deployment by name from the Kubernetes instance
func getDeploymentHandler(k8sConn comms.K8sCommunicator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		if name == "" {
			log.Println("name param required for this path")
			w.WriteHeader(http.StatusBadRequest)
		}
		deployment, err := k8sConn.GetDeployment(name)

		if err.Error() == comms.ErrDeploymentNotFound.Error() {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// TODO: catch other kinds of errors

		jsonData, err := deployment.SerializeForWire()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(jsonData)
	})
}

// updateDeploymentHandler updates a specific container on a single Deployment to a given image
func updateDeploymentHandler(k8sConn comms.K8sCommunicator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var deployment *comms.GzrDeployment
		name := mux.Vars(r)["name"]

		if name == "" {
			log.Println("name param required for this path")
			w.WriteHeader(http.StatusBadRequest)
		}
		deployment, err = k8sConn.GetDeployment(name)

		if err.Error() == comms.ErrDeploymentNotFound.Error() {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		userData := &UpdateDeploymentUserType{}

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(userData)

		if err != nil {
			log.Println("Error decoding JSON")
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		deployment, err = k8sConn.UpdateDeployment(userData.convertToDeploymentContainerInfo(k8sConn.GetNamespace(), name))

		// TODO: more fine-grained error reporting
		if err.Error() == comms.ErrContainerNotFound.Error() {
			log.Println(err)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jsonData, err := deployment.SerializeForWire()

		// TODO: more fine-grained error reporting
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.Write(jsonData)
	})
}

// convertToDeploymentContainerInfo creates a DeploymentContainerInfo struct
func (updateData *UpdateDeploymentUserType) convertToDeploymentContainerInfo(namespace string, deploymentName string) *comms.DeploymentContainerInfo {
	return &comms.DeploymentContainerInfo{
		Namespace:      namespace,
		DeploymentName: deploymentName,
		ContainerName:  updateData.ContainerName,
		Image:          updateData.Image,
	}
}
