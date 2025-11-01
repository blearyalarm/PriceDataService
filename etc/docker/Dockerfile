FROM golang:1.20-alpine as builder

WORKDIR /tmp/zeonology
COPY . /tmp/zeonology/

RUN apk update && apk add --no-cache make protobuf-dev
RUN protoc --version
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.1
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.26

RUN protoc --go_out=.gen/protos --proto_path=protos  --go_opt=paths=source_relative \
        --go-grpc_out=.gen/protos --go-grpc_opt=paths=source_relative,require_unimplemented_servers=false \
        $(find protos -iname "*.proto")

# and build a completely static binary (so we can use
# scratch as basis for the final image)
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on
RUN go build /tmp/zeonology/etc/script/generate_saml_keypair.go
RUN go build /tmp/zeonology/main.go

RUN GRPC_HEALTH_PROBE_VERSION=v0.3.1 && \
    wget -qO/bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/${GRPC_HEALTH_PROBE_VERSION}/grpc_health_probe-linux-amd64 && \
    chmod +x /bin/grpc_health_probe

FROM alpine:3.17.2

## ARG WorkingEnv
WORKDIR /app

COPY --from=builder /tmp/zeonology/main /app/main
COPY --from=builder /bin/grpc_health_probe /bin/grpc_health_probe
COPY --from=builder /tmp/zeonology/generate_saml_keypair /app/generate_saml_keypair

# Mark main as an executable file
RUN chmod +x /app/generate_saml_keypair
RUN chmod +x /app/main

EXPOSE 5051
EXPOSE 7070

ENTRYPOINT ./generate_saml_keypair && ./main
