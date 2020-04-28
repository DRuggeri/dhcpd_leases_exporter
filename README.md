# ISC dhcpd Leases Prometheus Exporter

A [Prometheus](https://prometheus.io) exporter for the ISC dhcpd server. This exporter consumes the `dhcpd.leases` file ([man page](https://linux.die.net/man/5/dhcpd.leases)) which is periodically written by the dameon. It is based on the [node_exporter](https://github.com/prometheus/node_exporter) and [cf_exporter](https://github.com/bosh-prometheus/cf_exporter) projects.

## Installation

### Binaries

Download the already existing [binaries](https://github.com/DRuggeri/dhcpd_leases_exporter/releases) for your platform:

```bash
$ ./dhcpd_leases_exporter <flags>
```

### From source

Using the standard `go install` (you must have [Go](https://golang.org/) already installed in your local machine):

```bash
$ go install github.com/DRuggeri/dhcpd_leases_exporter
$ dhcpd_leases_exporter <flags>
```

### With Docker
```bash
docker build -t dhcpd_leases_exporter .
docker run -d -p 9198:9198 -v /var/lib/dhcp:/var/lib/dhcpd:ro dhcpd_leases_exporter"
```

## Usage

### Flags

```
usage: dhcpd_leases_exporter [<flags>]

Flags:
  -h, --help              Show context-sensitive help (also try --help-long and --help-man).
      --dhcpd.leases="/var/lib/dhcp/dhcpd.leases"
                          Path of the dhcpd.leases file. Defaults to '/var/lib/dhcp/dhcpd.leases' ($DHCPD_LEASES_EXPORTER_LEASES_FILE)
      --filter.collectors="Stats"
                          Comma separated collectors to enable (Stats) ($DHCPD_LEASES_EXPORTER_FILTER_COLLECTORS)
      --metrics.namespace="dhcpd_leases"
                          Metrics Namespace ($DHCPD_LEASES_EXPORTER_METRICS_NAMESPACE)
      --web.listen-address=":9198"
                          Address to listen on for web interface and telemetry ($DHCPD_LEASES_EXPORTER_WEB_LISTEN_ADDRESS)
      --web.telemetry-path="/metrics"
                          Path under which to expose Prometheus metrics ($DHCPD_LEASES_EXPORTER_WEB_TELEMETRY_PATH)
      --web.auth.username=WEB.AUTH.USERNAME
                          Username for web interface basic auth ($DHCPD_LEASES_EXPORTER_WEB_AUTH_USERNAME)
      --web.auth.password=WEB.AUTH.PASSWORD
                          Password for web interface basic auth ($DHCPD_LEASES_EXPORTER_WEB_AUTH_PASSWORD)
      --web.tls.cert_file=WEB.TLS.CERT_FILE
                          Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate
                          ($DHCPD_LEASES_EXPORTER_WEB_TLS_CERTFILE)
      --web.tls.key_file=WEB.TLS.KEY_FILE
                          Path to a file that contains the TLS private key (PEM format) ($DHCPD_LEASES_EXPORTER_WEB_TLS_KEYFILE)
      --printMetrics      Print the metrics this exporter exposes and exits. Default: false ($DHCPD_LEASES_EXPORTER_PRINT_METRICS)
      --log.level="info"  Only log messages with the given severity or above. Valid levels: [debug, info, warn, error, fatal]
      --log.format="logger:stderr"
                          Set the log target and format. Example: "logger:syslog?appname=bob&local=7" or "logger:stdout?json=true"
      --version           Show application version.
```

## Metrics

### Stats
This collector counts the number of leases PER UNIQUE IP found in dhcpd.leases. This means that if an IP was leased to `client X`, but is now leased to `client Y`, there will be only one entry counted (the most recent one). This is in line with how dhcpd reads the file.

```
  dhcpd_leases_stats_valid - The number of leases in dhcpd.leases that have not yet expired.
  dhcpd_leases_stats_expired - The number of leases in dhcpd.leases that have expired.
  dhcpd_leases_stats_count - The number of leases in dhcpd.leases
  dhcpd_leases_stats_scrapes_total - Total number of scrapes for stats.
  dhcpd_leases_stats_scrape_errors_total - Total number of scrapes errors for stats.
  dhcpd_leases_stats_last_scrape_error - Whether the last scrape of stats resulted in an error (1 for error, 0 for success).
  dhcpd_leases_stats_last_scrape_timestamp - Number of seconds since 1970 since last scrape of stat metrics.
```

## Contributing

Refer to the [contributing guidelines](https://github.com/DRuggeri/dhcpd_leases_exporter/blob/master/CONTRIBUTING.md).

## License

Apache License 2.0, see [LICENSE](https://github.com/DRuggeri/dhcpd_leases_exporter/blob/master/LICENSE).
