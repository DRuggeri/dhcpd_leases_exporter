package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/DRuggeri/dhcpd_leases_exporter/collectors"
	"github.com/DRuggeri/dhcpd_leases_exporter/filters"
	"github.com/DRuggeri/dhcpdleasesreader"
)

var Version = "testing"

var (
	dhcpdLeasesFile = kingpin.Flag(
		"dhcpd.leases", "Path of the dhcpd.leases file. Defaults to '/var/lib/dhcp/dhcpd.leases' ($DHCPD_LEASES_EXPORTER_LEASES_FILE)",
	).Envar("DHCPD_LEASES_EXPORTER_LEASES_FILE").Default("/var/lib/dhcp/dhcpd.leases").String()

	filterCollectors = kingpin.Flag(
		"filter.collectors", "Comma separated collectors to enable (Stats,Leases) ($DHCPD_LEASES_EXPORTER_FILTER_COLLECTORS)",
	).Envar("DHCPD_LEASES_EXPORTER_FILTER_COLLECTORS").Default("Stats").String()

	metricsNamespace = kingpin.Flag(
		"metrics.namespace", "Metrics Namespace ($DHCPD_LEASES_EXPORTER_METRICS_NAMESPACE)",
	).Envar("DHCPD_LEASES_EXPORTER_METRICS_NAMESPACE").Default("dhcpd_leases").String()

	listenAddress = kingpin.Flag(
		"web.listen-address", "Address to listen on for web interface and telemetry ($DHCPD_LEASES_EXPORTER_WEB_LISTEN_ADDRESS)",
	).Envar("DHCPD_LEASES_EXPORTER_WEB_LISTEN_ADDRESS").Default(":9198").String()

	metricsPath = kingpin.Flag(
		"web.telemetry-path", "Path under which to expose Prometheus metrics ($DHCPD_LEASES_EXPORTER_WEB_TELEMETRY_PATH)",
	).Envar("DHCPD_LEASES_EXPORTER_WEB_TELEMETRY_PATH").Default("/metrics").String()

	authUsername = kingpin.Flag(
		"web.auth.username", "Username for web interface basic auth ($DHCPD_LEASES_EXPORTER_WEB_AUTH_USERNAME)",
	).Envar("DHCPD_LEASES_EXPORTER_WEB_AUTH_USERNAME").String()
	authPassword = ""

	tlsCertFile = kingpin.Flag(
		"web.tls.cert_file", "Path to a file that contains the TLS certificate (PEM format). If the certificate is signed by a certificate authority, the file should be the concatenation of the server's certificate, any intermediates, and the CA's certificate ($DHCPD_LEASES_EXPORTER_WEB_TLS_CERTFILE)",
	).Envar("DHCPD_LEASES_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()

	tlsKeyFile = kingpin.Flag(
		"web.tls.key_file", "Path to a file that contains the TLS private key (PEM format) ($DHCPD_LEASES_EXPORTER_WEB_TLS_KEYFILE)",
	).Envar("DHCPD_LEASES_EXPORTER_WEB_TLS_KEYFILE").ExistingFile()

	dhcpdLeasesPrintMetrics = kingpin.Flag(
		"printMetrics", "Print the metrics this exporter exposes and exits. Default: false ($DHCPD_LEASES_EXPORTER_PRINT_METRICS)",
	).Envar("DHCPD_LEASES_EXPORTER_PRINT_METRICS").Default("false").Bool()
)

func init() {
	prometheus.MustRegister(version.NewCollector(*metricsNamespace))
}

type basicAuthHandler struct {
	handler  http.HandlerFunc
	username string
	password string
}

func (h *basicAuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	if !ok || username != h.username || password != h.password {
		log.Errorf("Invalid HTTP auth from `%s`", r.RemoteAddr)
		w.Header().Set("WWW-Authenticate", "Basic realm=\"metrics\"")
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}
	h.handler(w, r)
	return
}

func prometheusHandler() http.Handler {
	handler := promhttp.Handler()

	if *authUsername != "" && authPassword != "" {
		handler = &basicAuthHandler{
			handler:  promhttp.Handler().ServeHTTP,
			username: *authUsername,
			password: authPassword,
		}
	}

	return handler
}

func main() {
	log.AddFlags(kingpin.CommandLine)
	kingpin.Version(Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	var mux = &sync.Mutex{}
	info, err := dhcpdleasesreader.NewDhcpdInfo(*dhcpdLeasesFile, false)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if *dhcpdLeasesPrintMetrics {
		/* Make a channel and function to send output along */
		var out chan *prometheus.Desc
		eatOutput := func(in <-chan *prometheus.Desc) {
			for desc := range in {
				/* Weaksauce... no direct access to the variables */
				//Desc{fqName: "the_name", help: "help text", constLabels: {}, variableLabels: []}
				tmp := desc.String()
				vals := strings.Split(tmp, `"`)
				fmt.Printf("  %s - %s\n", vals[1], vals[3])
			}
		}

		/* Interesting juggle here...
		   - Make a channel the describe function can send output to
		   - Start the printing function that consumes the output in the background
		   - Call the describe function to feed the channel (which blocks until the consume function eats a message)
		   - When the describe function exits after returning the last item, close the channel to end the background consume function
		*/
		fmt.Println("Stats")
		statsCollector := collectors.NewStatsCollector(*metricsNamespace, info, mux)
		out = make(chan *prometheus.Desc)
		go eatOutput(out)
		statsCollector.Describe(out)
		close(out)

		fmt.Println("Leases")
		leasesCollector := collectors.NewLeaseCollector(*metricsNamespace, info, mux)
		out = make(chan *prometheus.Desc)
		go eatOutput(out)
		leasesCollector.Describe(out)
		close(out)

		os.Exit(0)
	}

	log.Infoln("Starting dhcpd_leases_exporter", Version)
	authPassword = os.Getenv("DHCPD_LEASES_EXPORTER_WEB_AUTH_PASSWORD")

	var collectorsFilters []string
	if *filterCollectors != "" {
		collectorsFilters = strings.Split(*filterCollectors, ",")
	}
	collectorsFilter, err := filters.NewCollectorsFilter(collectorsFilters)
	if err != nil {
		log.Error(err)
		os.Exit(1)
	}

	if collectorsFilter.Enabled(filters.StatsCollector) {
		statsCollector := collectors.NewStatsCollector(*metricsNamespace, info, mux)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		prometheus.MustRegister(statsCollector)
	}

	if collectorsFilter.Enabled(filters.LeaseCollector) {
		activeCollector := collectors.NewLeaseCollector(*metricsNamespace, info, mux)
		if err != nil {
			log.Error(err)
			os.Exit(1)
		}
		prometheus.MustRegister(activeCollector)
	}

	handler := prometheusHandler()
	http.Handle(*metricsPath, handler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>DHCPD Leases Exporter</title></head>
             <body>
             <h1>DHCPD Leases Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})

	if *tlsCertFile != "" && *tlsKeyFile != "" {
		log.Infoln("Listening TLS on", *listenAddress)
		log.Fatal(http.ListenAndServeTLS(*listenAddress, *tlsCertFile, *tlsKeyFile, nil))
	} else {
		log.Infoln("Listening on", *listenAddress)
		log.Fatal(http.ListenAndServe(*listenAddress, nil))
	}
}
