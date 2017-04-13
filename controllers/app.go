package controllers

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/bypasslane/gzr/comms"
	"github.com/bypasslane/gzr/middleware"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/urfave/negroni"
	"github.com/pkg/errors"

	"fmt"
)


func App(k8sConn comms.K8sCommunicator, imageStore comms.GzrMetadataStore) http.Handler {
	router := mux.NewRouter().StrictSlash(true).UseEncodedPath()

	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/deployments", listDeploymentsHandler(k8sConn)).Methods("GET")
	router.HandleFunc("/deployments/{name}", getDeploymentHandler(k8sConn)).Methods("GET")
	router.HandleFunc("/deployments/{name}", updateDeploymentHandler(k8sConn)).Methods("PUT")

	router.HandleFunc("/images/{name}", getImagesHandler(imageStore)).Methods("GET")
	router.HandleFunc("/images/{name}/{version}", getImageHandler(imageStore)).Methods("GET")

	//middleware setup (basically same as classic but uses our logrus for logging)

	recovery := negroni.NewRecovery()
	loggerMiddleware := negronilogrus.NewCustomMiddleware(log.GetLevel(), &log.JSONFormatter{}, "web")

	static := negroni.NewStatic(http.Dir("public"))
	jsonHeader := middleware.NewContentType()

	n := negroni.New(recovery, loggerMiddleware, static, jsonHeader)

	n.UseHandler(router)

	return n
}

// homeHandler handles requests to the root of the server
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("CHOOSE THE FORM"))
}

type stackTracer interface {
	StackTrace() errors.StackTrace
}

func logErrorFields(err error) *log.Entry{
	logEntry :=  log.WithError(err)
	if err, ok := err.(stackTracer); ok {
		logEntry = logEntry.WithField("stacktrace", fmt.Sprintf("%+v", err.(stackTracer).StackTrace()))
	}
	return logEntry
}