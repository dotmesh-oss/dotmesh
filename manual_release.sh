#!/bin/bash

# For when times are hard, and you don't have CI wired up yet.
# Log into quay.io first.

export REPOSITORY=quay.io/dotmesh/
export DOCKER_TAG=release-0.8.2
export VERSION=release-0.8.2

make release_all
