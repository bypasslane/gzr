package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bypasslane/gzr/comms"

	log "github.com/Sirupsen/logrus"
	"github.com/ericchiang/k8s/apis/extensions/v1beta1"
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
func listDeploymentsHandler(k8sComm comms.K8sCommunicator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		deployments, err := k8sComm.ListDeployments()
		// TODO: differentiate between legit errors and unhandleable errors
		if err != nil {
			logErrorFields(err).Error("Unable to list deployments")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		if deployments == nil {
			log.Warn("No deployments found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jsonData, err := comms.SerializeForWire(deployments)

		if err != nil {
			logErrorFields(err).Error("Error serializing for wire")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(jsonData)
	})
}

// getDeploymentHandler gets a single Deployment by name from the Kubernetes instance
func getDeploymentHandler(k8sComm comms.K8sCommunicator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]

		if name == "" {
			log.Warn("name param required for this path")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("name param required for this path"))
		}
		deployment, err := k8sComm.GetDeployment(name)

		if errors.Cause(err) == comms.ErrDeploymentNotFound {
			logErrorFields(err).Warnf("Deployment not found for %q", name)
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		// TODO: catch other kinds of errors

		jsonData, err := comms.SerializeForWire(deployment)

		if err != nil {
			logErrorFields(err).Error("Error serializng for wire")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(jsonData)
	})
}

// updateDeploymentHandler updates a specific container on a single Deployment
func updateDeploymentHandler(k8sComm comms.K8sCommunicator) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var deployment *v1beta1.Deployment
		var foundContainer bool
		name := mux.Vars(r)["name"]

		if name == "" {
			log.Warn("name param required for this path")
			w.WriteHeader(http.StatusBadRequest)
		}
		deployment, err = k8sComm.GetDeployment(name)

		if errors.Cause(err) == comms.ErrDeploymentNotFound {
			logErrorFields(err).Warn("Error getting deployment")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(err.Error()))
			return
		}

		userData := &UpdateDeploymentUserType{}

		decoder := json.NewDecoder(r.Body)
		err = decoder.Decode(userData)

		if err != nil {
			logErrorFields(err).Warn("Error decoding JSON")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		for containerIndex, container := range deployment.Spec.Template.Spec.Containers {
			if *container.Name == userData.ContainerName {
				foundContainer = true
				*deployment.Spec.Template.Spec.Containers[containerIndex].Image = userData.Image
				break
			}
		}
		if !foundContainer {
			logErrorFields(err).Warn("Could not find container with name %q", userData.ContainerName)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(fmt.Sprintf("Could not find container with name %q", userData.ContainerName)))
		}

		deployment, err = k8sComm.UpdateDeployment(deployment)

		jsonData, err := comms.SerializeForWire(deployment)

		// TODO: more fine-grained error reporting
		if err != nil {
			logErrorFields(err).Error("Error serialzing for wire")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(jsonData)
	})
}
