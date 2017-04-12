package controllers

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"

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
			log.WithError(err).Error("Unable to list deployments")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if deployments == nil {
			log.Warn("No deployments found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jsonData, err := deployments.SerializeForWire()

		if err != nil {
			log.WithError(err).Error("Error serializing for wire")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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
			log.Warn("name param required for this path")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("name param required for this path"))
		}
		deployment, err := k8sConn.GetDeployment(name)

		if errors.Cause(err) == comms.ErrDeploymentNotFound {
			log.WithError(err).Warn("Deployment not found")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		// TODO: catch other kinds of errors

		jsonData, err := deployment.SerializeForWire()

		if err != nil {
			log.WithError(err).Error("Error serializng for wire")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
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
			log.Warn("name param required for this path")
			w.WriteHeader(http.StatusBadRequest)
		}
		deployment, err = k8sConn.GetDeployment(name)

		if errors.Cause(err) == comms.ErrDeploymentNotFound {
			log.WithError(err).Warn("Error getting deployment")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		userData := &UpdateDeploymentUserType{}

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(userData)

		if err != nil {
			log.WithError(err).Warn("Error decoding JSON")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		deployment, err = k8sConn.UpdateDeployment(userData.convertToDeploymentContainerInfo(k8sConn.GetNamespace(), name))

		// TODO: more fine-grained error reporting
		if errors.Cause(err) == comms.ErrContainerNotFound {
			log.WithError(err).Warn("Conatiner not found")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		jsonData, err := deployment.SerializeForWire()

		// TODO: more fine-grained error reporting
		if err != nil {
			log.WithError(err).Error("Error serialzing for wire")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
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
