package collectors

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"github.com/DRuggeri/dhcpdleasesreader"
)

type StatCollector struct {
	namespace   string
	validMetric prometheus.Gauge
	expiredMetric prometheus.Gauge
	countMetric prometheus.Gauge
	modTimeMetric prometheus.Gauge
	info       *dhcpdleasesreader.DhcpdInfo

	scrapesTotalMetric              prometheus.Counter
	scrapeErrorsTotalMetric         prometheus.Counter
	lastScrapeErrorMetric           prometheus.Gauge
	lastScrapeTimestampMetric       prometheus.Gauge
	lastScrapeDurationSecondsMetric prometheus.Gauge
}

func NewStatsCollector(namespace string, info *dhcpdleasesreader.DhcpdInfo) *StatCollector {
	validMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "valid",
			Help:      "The number of leases in dhcpd.leases that have not yet expired.",
		},
	)

	expiredMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "expired",
			Help:      "The number of leases in dhcpd.leases that have expired.",
		},
	)

	countMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "count",
			Help:      "The number of leases in dhcpd.leases",
		},
	)

	modTimeMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "filetime",
			Help:      "The file timestamp in seconds since epoch of the dhcpd.leases file",
		},
	)


	scrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "scrapes_total",
			Help:      "Total number of scrapes for stats.",
		},
	)

	scrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "scrape_errors_total",
			Help:      "Total number of scrapes errors for stats.",
		},
	)

	lastScrapeErrorMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "last_scrape_error",
			Help:      "Whether the last scrape of stats resulted in an error (1 for error, 0 for success).",
		},
	)

	lastScrapeTimestampMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "last_scrape_timestamp",
			Help:      "Number of seconds since 1970 since last scrape of stat metrics.",
		},
	)

	lastScrapeDurationSecondsMetric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "last_scrape_duration_seconds",
			Help:      "Duration of the last scrape of stats.",
		},
	)

	return &StatCollector{
		validMetric: validMetric,
		expiredMetric: expiredMetric,
		countMetric: countMetric,
		modTimeMetric: modTimeMetric,
		info: info,

		namespace:   namespace,
		scrapesTotalMetric:              scrapesTotalMetric,
		scrapeErrorsTotalMetric:         scrapeErrorsTotalMetric,
		lastScrapeErrorMetric:           lastScrapeErrorMetric,
		lastScrapeTimestampMetric:       lastScrapeTimestampMetric,
		lastScrapeDurationSecondsMetric: lastScrapeDurationSecondsMetric,
	}
}

func (c *StatCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()

	c.info.Read()

	c.validMetric.Set(float64(c.info.Valid))
	c.validMetric.Collect(ch)

	c.expiredMetric.Set(float64(c.info.Expired))
	c.expiredMetric.Collect(ch)

	c.countMetric.Set(float64(len(c.info.Leases)))
	c.countMetric.Collect(ch)

	c.modTimeMetric.Set(float64(c.info.ModTime.Unix()))
	c.modTimeMetric.Collect(ch)

	c.scrapeErrorsTotalMetric.Collect(ch)

	c.scrapesTotalMetric.Inc()
	c.scrapesTotalMetric.Collect(ch)

	c.lastScrapeErrorMetric.Set(0)
	c.lastScrapeErrorMetric.Collect(ch)

	c.lastScrapeTimestampMetric.Set(float64(time.Now().Unix()))
	c.lastScrapeTimestampMetric.Collect(ch)

	c.lastScrapeDurationSecondsMetric.Set(time.Since(begun).Seconds())
	c.lastScrapeDurationSecondsMetric.Collect(ch)
}

func (c *StatCollector) Describe(ch chan<- *prometheus.Desc) {
	c.validMetric.Describe(ch)
	c.expiredMetric.Describe(ch)
	c.countMetric.Describe(ch)
	c.scrapesTotalMetric.Describe(ch)
	c.scrapeErrorsTotalMetric.Describe(ch)
	c.lastScrapeErrorMetric.Describe(ch)
	c.lastScrapeTimestampMetric.Describe(ch)
	c.lastScrapeDurationSecondsMetric.Describe(ch)
}
