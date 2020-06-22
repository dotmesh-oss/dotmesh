FROM golang:1.12.5 AS build-env
WORKDIR /usr/local/go/src/github.com/dotmesh-oss/dotmesh
COPY ./cmd /usr/local/go/src/github.com/dotmesh-oss/dotmesh/cmd
COPY ./pkg /usr/local/go/src/github.com/dotmesh-oss/dotmesh/pkg
COPY ./vendor /usr/local/go/src/github.com/dotmesh-oss/dotmesh/vendor
RUN cd cmd/dynamic-provisioner && go install -ldflags '-linkmode external -extldflags "-static"'

FROM scratch
COPY --from=build-env /usr/local/go/bin/dynamic-provisioner /
CMD ["/dynamic-provisioner"]