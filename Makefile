build_client:
	mkdir -p binaries/Linux && CGO_ENABLED=0 go build -ldflags "-w -s -X main.clientVersion=${VERSION} -X main.dockerTag=${DOCKER_TAG}" -o binaries/Linux/dm ./cmd/dm

build_client_mac:
	mkdir -p binaries/Darwin && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags "-w -s -X main.clientVersion=${VERSION} -X main.dockerTag=${DOCKER_TAG}" -o binaries/Darwin/dm ./cmd/dm

# if you want to use a repository, it must end in `/`
build_server: 
	docker build -t ${REPOSITORY}dotmesh-server:${DOCKER_TAG} . --build-arg VERSION=${VERSION} --build-arg STABLE_DOCKER_TAG=${DOCKER_TAG} -f dockerfiles/dotmesh.Dockerfile

build_dind_prov: 
	docker build -t ${REPOSITORY}dind-dynamic-provisioner:${DOCKER_TAG} . -f dockerfiles/dind-provisioner.Dockerfile

push_server: 
	docker push ${REPOSITORY}dotmesh-server:${DOCKER_TAG}

push_dind_prov:
	docker push ${REPOSITORY}dind-dynamic-provisioner:${DOCKER_TAG}

build_operator:
	docker build -t ${REPOSITORY}dotmesh-operator:${DOCKER_TAG} --build-arg VERSION=${VERSION} --build-arg STABLE_DOTMESH_SERVER_IMAGE=${REPOSITORY}dotmesh-server:${DOCKER_TAG} . -f dockerfiles/operator.Dockerfile

push_operator:
	docker push ${REPOSITORY}dotmesh-operator:${DOCKER_TAG}

build_provisioner: 
	docker build -t ${REPOSITORY}dotmesh-dynamic-provisioner:${DOCKER_TAG} . -f dockerfiles/provisioner.Dockerfile

push_provisioner: 
	docker push ${REPOSITORY}dotmesh-dynamic-provisioner:${DOCKER_TAG}

gitlab_registry_login: 
	docker login -u gitlab-ci-token -p ${CI_BUILD_TOKEN} ${CI_REGISTRY}

release_all: 
	make build_server && make build_operator && make build_provisioner && make push_provisioner && make push_server && make push_operator

rebuild:
	make build_server && make build_operator && make build_provisioner

build_push_server:
	make gitlab_registry_login && make build_server && make build_dind_prov && make push_server && make push_dind_prov

build_push_operator:
	make gitlab_registry_login && make build_operator && make push_operator

build_push_provisioner:
	make gitlab_registry_login && make build_provisioner && make push_provisioner

prep_tests:
	./scripts/mark-cleanup.sh && make clear_context && export DOCKER_TAG=latest && export VERSION=latest && make rebuild

reset_bucket:
	docker run -it --env S3_ACCESS_KEY=${S3_ACCESS_KEY} --env S3_SECRET_KEY=${S3_SECRET_KEY} -v $(dir $(realpath $(firstword $(MAKEFILE_LIST))))/scripts:/scripts garland/docker-s3cmd /scripts/create-s3-stress-test-bucket.sh
