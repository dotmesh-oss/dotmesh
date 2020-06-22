FROM golang:1.12.5 AS build-env
ARG VERSION
ARG STABLE_DOTMESH_SERVER_IMAGE
WORKDIR /usr/local/go/src/github.com/dotmesh-oss/dotmesh
COPY ./cmd /usr/local/go/src/github.com/dotmesh-oss/dotmesh/cmd
COPY ./pkg /usr/local/go/src/github.com/dotmesh-oss/dotmesh/pkg
COPY ./vendor /usr/local/go/src/github.com/dotmesh-oss/dotmesh/vendor
RUN cd cmd/operator && go install -ldflags "-linkmode external -extldflags '-static' -X main.DOTMESH_VERSION=${VERSION} -X main.DOTMESH_IMAGE=${STABLE_DOTMESH_SERVER_IMAGE}"

FROM scratch
COPY --from=build-env /usr/local/go/bin/operator /
CMD ["/operator"]