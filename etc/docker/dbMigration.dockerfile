FROM alpine:3.16.1

RUN GOLANG_MIGRATE_VERSION=v4.15.2 && \
    wget https://github.com/golang-migrate/migrate/releases/download/${GOLANG_MIGRATE_VERSION}/migrate.linux-amd64.tar.gz -O - |  tar -xz -C /bin && \
    chmod +x /bin/migrate

WORKDIR /app

COPY ./migration /app/migration

ENTRYPOINT ["/bin/migrate"]
