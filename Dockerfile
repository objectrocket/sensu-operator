# build stage
FROM golang:1.10 AS build-env
ARG APPVERSION=latest
WORKDIR /usr/local/src/sensu-operator
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-X github.com/objectrocket/sensu-operator/version.Version=$APPVERSION" -ldflags="-X google.golang.org/protobuf/reflect/protoregistry.conflictPolicy=warn" -o _output/sensu-operator cmd/operator/main.go

FROM alpine:3.6
ENV USER=sensu-operator
COPY --from=build-env /usr/local/src/sensu-operator/_output/sensu-operator /usr/local/bin/sensu-operator
RUN apk add --no-cache --update ca-certificates && \
    addgroup -g 1000 ${USER} && \
    adduser -D -g "${USER} user" -H -h "/app" -G "${USER}" -u 1000 ${USER} && \
    chown -R ${USER}:${USER} /usr/local/bin/sensu-operator
USER ${USER}:${USER}
ENTRYPOINT ["./usr/local/bin/sensu-operator"]
