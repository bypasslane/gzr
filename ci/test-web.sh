#!/usr/bin/env bash

set -eu

git config --global url."https://".insteadOf git://
git config --global url."https://".insteadOf ssh://
git config --global url."https://github.com/".insteadOf git@github.com:

pushd gozer-web
    # tests don't exist yet, uncomment when they do @JonathanTech
    npm install
    #npm test
popd
