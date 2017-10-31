#!/usr/bin/env bash

set -eu

go get github.com/Masterminds/glide

make install
make test
