FROM golang:1.12.5-alpine3.9 AS build-env
WORKDIR /usr/local/go/src/github.com/dotmesh-io/dotmesh
RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh
COPY ./cmd /usr/local/go/src/github.com/dotmesh-io/dotmesh/cmd
COPY ./pkg /usr/local/go/src/github.com/dotmesh-io/dotmesh/pkg
COPY ./vendor /usr/local/go/src/github.com/dotmesh-io/dotmesh/vendor
RUN cd cmd/dotmesh-server/pkg/dind-dynamic-provisioning && go install -ldflags '-linkmode external -extldflags "-static"'

FROM scratch
COPY --from=build-env /usr/local/go/bin/dind-dynamic-provisioning /
CMD ["/dind-dynamic-provisioning"]