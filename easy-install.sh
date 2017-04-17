#!/usr/bin/env bash

set -eu

initArch() {
    ARCH=$(uname -m)
    case $ARCH in
        armv5*) ARCH="armv5";;
        armv6*) ARCH="armv6";;
        armv7*) ARCH="armv7";;
        aarch64) ARCH="arm64";;
        x86) ARCH="386";;
        x86_64) ARCH="amd64";;
        i686) ARCH="386";;
        i386) ARCH="386";;
    esac
}

initOS() {
    OS=$(echo `uname`|tr '[:upper:]' '[:lower:]')

    case "$OS" in
        # Minimalist GNU for Windows
        mingw*) OS='windows';;
    esac
}

installRelease() {
    F="gzr-${OS}-${ARCH}"
    RELEASE=0
    if wget -q https://github.com/bypasslane/gzr/releases/download/latest/$F -O /tmp &> /dev/null; then
        tar -xvf /tmp/$F -C .
        echo "gzr executable downloaded and placed at $PWD/gzr"
    else
        echo "could not find release artifact $F at https://github.com/bypasslane/gzr/releases/latest"
        RELEASE=1
    fi
}

masterDeps() {
    req="git, zip, go, and glide"
    if ! hash git 2>/dev/null; then
        echo "please install required: ${req}"
        exit 1
    fi
    if ! hash zip 2>/dev/null; then
        echo "please install required: ${req}"
        exit 1
    fi
    if ! hash go 2>/dev/null; then
        echo "please install required: ${req}"
        exit 1
    fi
    if ! hash glide 2>/dev/null; then
        echo "please install required: ${req}"
        exit 1
    fi
    return 0
}

installMaster() {
    echo "attempting to build master and install locally"
    masterDeps
    rm -rf /tmp/gzr
    if git clone git@github.com:bypasslane/gzr.git /tmp/gzr; then
        mkdir -p $GOPATH/src/github.com/bypasslane
        rm -rf $GOPATH/src/github.com/bypasslane/gzr
        mv /tmp/gzr $GOPATH/src/github.com/bypasslane
    else
        echo "failed to clone git@github.com:bypasslane/gzr.git"
        exit 1
    fi

    pushd $GOPATH/src/github.com/bypasslane/gzr
        if ! make install_deps; then
            echo "failed to install dependencies"
            exit 1
        fi
        if ! make install; then
            echo "failed glide install"
            exit 1
        fi
        if ! make build; then
            echo "failed glide build"
            exit 1
        fi
    popd

    cp $GOPATH/src/github.com/bypasslane/gzr/gzr .
    echo "gzr executable downloaded and placed at $PWD/gzr"
}

initArch
initOS
installRelease
if [ "$RELEASE" -eq 1 ]; then
    installMaster
fi
