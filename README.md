# gzr

*"Gozer"*

The missing (web or CLI) tool for your (containerized) microservices SDLC

## What's it do?

* Lets you view your microservices as a collection of running code projects
* Provides web and command line interfaces to update deployments and to know what code built an image
* Stores metadata about built container images. Supports etcd and BoltDB.
* Tracks container metadata to provide information that registries don't always have:
	* CI build information
	* issue tracker information
  * originating repository
  * easy link to last 10 closed PRs

## Who's it for?
* Ops/QA
* Continual Integration Systems
* Developers working on 1 microservice with N microservices supporting

## Assumptions
* Git is the SCM
* GitHub is the repository origin
* [Kubernetes](https://kubernetes.io) (k8S) is the orchestration tool
* [Docker](https://www.docker.com) is the container runtime
* `$HOME/.kube/config` holds a k8s live configuration

## Why?
We needed an easy way to manage deployments in a variety of contexts, thinking in terms of "built repo branches" instead of "container with SHA of X". We didn't want to have to teach Kubernetes to everyone in our org who needs to deploy containers. We decided to build our own **coarser-grained** tool for deploying that would also be able to handle any metadata annotations we found it useful to make to our container images.

## Usage
`gzr help` shows you what you need to know for CLI usage

`gzr web` stands up the web interface - See [Gozer Web Docs](https://github.com/bypasslane/gzr/blob/master/gozer-web/README.md)

`gzr build` both builds a Docker image and pushes it to your repository and gzr's metadata store.


## Development

### Dependencies
* Node version > 4
* rice
* zip
* git
* go
* glide
* kubectl
 
### Structure
`gzr` is a CLI tool written with Cobra. It has a `web` 
command that stands up a web UI based on [Gorilla](http://www.gorillatoolkit.org), [Negroni](https://github.com/urfave/negroni), 
[Rice](https://github.com/GeertJohan/go.rice), [Twitter Bootstrap](http://getbootstrap.com), and [Vue.js](https://vuejs.org/). 
The web handlers and CLI handlers both use the same `comms` package to talk to k8s and storage backends.

### Example configs and data
`image.example.json` contains an example of the image metadata expected for `store` commands. `.gzr.bolt.json` and `.gzr.etcd.json` contain example configuration files for each of those storage backends.
You can load the sample data with `make build && ./gzr image store test:1.0 $(pwd)/image.example.json`.
Your config file should be stored in $HOME/.gzr.json. If you are using the BoltDB backend the path supplied in this file must consist of existing directories.

### Make commands
* `make` and `make build` - builds gzr executable
* `make build_web` - builds web assets and uses rice tool to append them to executable
* `make install_build_deps` - installs rice cli tool

### Project folders

* **cmd** - a bunch of Cobra commands and some utilities
* **comms** - package for talking to k8s and storage backends
* **controllers** - the web app routes and handlers
* **gozer-web** - The vue.js web frontend. 