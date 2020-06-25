#!/usr/bin/env bash
set -xe
./scripts/mark-cleanup.sh
export VERSION=latest
export DOCKER_TAG=latest
export CI_REGISTRY="$(hostname).local:80"
export CI_REPOSITORY="dotmesh"
export REPOSITORY="${CI_REGISTRY}/${CI_REPOSITORY}/"
make build_server && make push_server
make build_operator && make push_operator
make build_provisioner && make push_provisioner
make build_client
make build_dind_prov && make push_dind_prov
make build_dind_flexvolume
