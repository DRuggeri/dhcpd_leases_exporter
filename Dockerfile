### STAGE 1: Build ###

FROM golang:buster as builder

RUN mkdir -p /app/src/github.com/DRuggeri/dhcpd_leases_exporter
ENV GOPATH /app
WORKDIR /app
COPY . /app/src/github.com/DRuggeri/dhcpd_leases_exporter
RUN go install github.com/DRuggeri/dhcpd_leases_exporter

### STAGE 2: Setup ###

FROM alpine
RUN apk add --no-cache \
  libc6-compat
COPY --from=builder /app/bin/dhcpd_leases_exporter /dhcpd_leases_exporter
RUN chmod +x /dhcpd_leases_exporter
ENTRYPOINT ["/dhcpd_leases_exporter"]
