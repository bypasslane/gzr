package controllers

import (
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/bypasslane/boxedRice"
	"github.com/bypasslane/gzr/comms"
	"github.com/bypasslane/gzr/middleware"
	"github.com/gorilla/mux"
	"github.com/meatballhat/negroni-logrus"
	"github.com/pkg/errors"
	"github.com/urfave/negroni"
)

// Allows for dependency injection of boxedRice.Config,
// preventing errors during tests when a public folder isn't found
type staticFileBoxConfig interface {
	MustFindBox(boxName string) *boxedRice.Box
}

// App builds our web router/handler
func App(k8sComm comms.K8sCommunicator, imageStore comms.GzrMetadataStore, boxedRiceConfig staticFileBoxConfig) http.Handler {
	router := mux.NewRouter().StrictSlash(true).UseEncodedPath()

	router.HandleFunc("/", homeHandler).Methods("GET")
	router.HandleFunc("/deployments", listDeploymentsHandler(k8sComm)).Methods("GET")
	router.HandleFunc("/deployments/{name}", getDeploymentHandler(k8sComm)).Methods("GET")
	router.HandleFunc("/deployments/{name}", updateDeploymentHandler(k8sComm)).Methods("PUT")

	router.HandleFunc("/images/{name}", getImagesHandler(imageStore)).Methods("GET")
	router.HandleFunc("/images/{name}/{version}", getImageHandler(imageStore)).Methods("GET")

	//middleware setup (basically same as classic but uses our logrus for logging)

	recovery := negroni.NewRecovery()
	recovery.Logger = log.StandardLogger()

	loggerMiddleware := negronilogrus.NewMiddlewareFromLogger(log.StandardLogger(), "web")

	static := negroni.NewStatic(boxedRiceConfig.MustFindBox("public").HTTPBox())
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

func logErrorFields(err error) *log.Entry {
	logEntry := log.WithError(err)
	if err, ok := err.(stackTracer); ok {
		logEntry = logEntry.WithField("stacktrace", fmt.Sprintf("%+v", err.(stackTracer).StackTrace()))
	}
	return logEntry
}
