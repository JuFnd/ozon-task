FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY . .

RUN go build -o authorization cmd/authorization/main.go
RUN go build -o posts cmd/posts/sever.go

FROM ubuntu:latest

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get -y install postgresql postgresql-contrib

USER postgres

COPY database /opt/database
RUN service postgresql start && \
        psql -c "CREATE USER boss WITH superuser login password 'boss';" && \
        psql -c "ALTER ROLE boss WITH PASSWORD 'boss';" && \
        createdb -O boss auth_service && \
        createdb -O boss films_service && \
        psql -d auth_service -f /opt/database/auth_service_migrations.sql && \
        psql -d posts_service -f /opt/database/films_service_migrations.sql

VOLUME ["/etc/postgresql", "/var/log/postgresql", "/var/lib/postgresql"]

USER root

WORKDIR /build
COPY --from=builder /app/configs .
COPY --from=builder /app/authorization .
COPY --from=builder /app/posts .

COPY . .

CMD service postgresql start \
    nohup ./authorization > /dev/null 2>&1 & \
    nohup ./posts > /dev/null 2>&1 &