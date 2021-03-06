#!/bin/bash

# For when times are hard, and you don't have CI wired up yet.
# Log into quay.io first.

export REPOSITORY=quay.io/dotmesh/
export DOCKER_TAG=$(git describe --tags)
export VERSION=$(git describe --tags)

make release_all
