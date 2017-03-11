#!/bin/bash

export PROJECTTOP=$(pwd)
export PROJECTROOT=${PROJECTTOP/\/src\/ScreenStreamer/}
export GOBIN=$PROJECTTOP/bin
export GOPATH="$PROJECTROOT:$PROJECTTOP/lib"