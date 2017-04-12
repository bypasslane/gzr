package controllers

import (
	"fmt"

	"net/http"
	"net/url"

	log "github.com/Sirupsen/logrus"
	"github.com/bypasslane/gzr/comms"
	"github.com/gorilla/mux"
	"github.com/spf13/viper"
)

func getImagesHandler(imageStore comms.GzrMetadataStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		name, err := url.QueryUnescape(name)
		if name == "" {
			log.Warn("name param required for this path")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("name param required for this path"))
			return
		}

		name = fmt.Sprintf("%s/%s", viper.GetString("repository"), name)
		images, err := imageStore.List(name)
		if err != nil {
			log.WithError(err).Warnf("Error retrieving images for %q", name)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		if len(images.Images) == 0 {
			log.Warnf("Images not found for %q", name)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jsonData, err := images.SerializeForWire()
		if err != nil {
			log.WithError(err).Error("Error serializing images")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(jsonData)
	})
}

func getImageHandler(imageStore comms.GzrMetadataStore) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := mux.Vars(r)["name"]
		name, err := url.QueryUnescape(name)
		if err != nil {
			log.WithError(err).Warn("name parameter in unexpected format")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		version := mux.Vars(r)["version"]
		if name == "" || version == "" {
			log.Warn("name and version required for this path")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("name and version required for this path"))
			return
		}
		searchString :=  fmt.Sprintf("%s:%s", name, version)
		image, err := imageStore.Get(searchString)
		if err != nil {
			log.WithError(err).Warn("image store failed to retrieve value for %q", searchString)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		// If empty Image, one wasn't found
		if image == nil {
			log.Warn("image not found")
			w.WriteHeader(http.StatusNotFound)
			return
		}

		jsonData, err := image.SerializeForWire()
		if err != nil {
			log.WithError(err).Error("Error serializng image data for wire")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(jsonData)
	})
}
