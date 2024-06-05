FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY ../.. .

RUN go build -o posts ./cmd/posts/server.go

FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

COPY scripts /opt/scripts

RUN sh /opt/scripts/redis.sh

RUN apt-get update && apt-get -y install postgresql postgresql-contrib

USER postgres

COPY database /opt/database
RUN service postgresql start && \
        psql -c "CREATE USER boss WITH superuser login password 'boss';" && \
        psql -c "ALTER ROLE boss WITH PASSWORD 'boss';" && \
        createdb -O boss posts_service && \
        psql -d posts_service -f /opt/database/posts_service_migrations.sql

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

WORKDIR /build
COPY --from=builder /app/configs .
COPY --from=builder /app/posts .

COPY . .

EXPOSE 8081

CMD service redis-server start && service postgresql start && ./posts