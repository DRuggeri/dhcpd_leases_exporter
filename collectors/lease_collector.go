package collectors

import (
	"github.com/DRuggeri/dhcpdleasesreader"
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type leaseCollector struct {
	activeDesc   *prometheus.Desc
	info        *dhcpdleasesreader.DhcpdInfo

	scrapesTotalMetric      prometheus.Counter
	scrapeErrorsTotalMetric prometheus.Counter
	lastScrapeErrorDesc     *prometheus.Desc
	lastScrapeTimestampDesc *prometheus.Desc
	lastScrapeDurationDesc  *prometheus.Desc
}

func NewLeaseCollector(namespace string, info *dhcpdleasesreader.DhcpdInfo) *leaseCollector {
	activeDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "active", "client"),
		"The number of leases in dhcpd.leases that have not yet expired",
		[]string{"hostname", "ip", "mac"}, nil,
	)

	scrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "active",
			Name:      "scrapes_total",
			Help:      "Total number of scrapes",
		},
	)

	scrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "active",
			Name:      "scrape_errors_total",
			Help:      "Total number of scrapes errors",
		},
	)

	lastScrapeErrorDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "active", "last_scrape_error"),
		"Whether the last scrape resulted in an error (1 for error, 0 for success).",
		nil, nil,
	)

	lastScrapeTimestampDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "active", "last_scrape_timestamp"),
		"Number of seconds since 1970 since last scrape.",
		nil, nil,
	)

	lastScrapeDurationDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "active", "last_scrape_duration_seconds"),
		"Number of seconds the last scrape took",
		nil, nil,
	)

	return &leaseCollector{
		activeDesc:   activeDesc,
                info:        info,

		scrapesTotalMetric:      scrapesTotalMetric,
		scrapeErrorsTotalMetric: scrapeErrorsTotalMetric,
		lastScrapeErrorDesc:     lastScrapeErrorDesc,
		lastScrapeTimestampDesc: lastScrapeTimestampDesc,
		lastScrapeDurationDesc:  lastScrapeDurationDesc,
	}
}

func (c *leaseCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()
	err_num := 0

	/* TODO: Surface read errors through this function */
	c.info.Read()

	c.scrapesTotalMetric.Inc()
	if err_num != 0 {
		c.scrapeErrorsTotalMetric.Inc()
	}

	for address, lease := range c.info.Leases {
		if begun.Before(lease.Ends) {
			ch <- prometheus.MustNewConstMetric(c.activeDesc, prometheus.GaugeValue, float64(1), lease.Hostname, address, lease.Hardware_address)
		}
	}

	ch <- c.scrapesTotalMetric
	ch <- c.scrapeErrorsTotalMetric
	ch <- prometheus.MustNewConstMetric(c.lastScrapeErrorDesc, prometheus.GaugeValue, float64(err_num))
	ch <- prometheus.MustNewConstMetric(c.lastScrapeTimestampDesc, prometheus.GaugeValue, float64(time.Now().Unix()))
	ch <- prometheus.MustNewConstMetric(c.lastScrapeDurationDesc, prometheus.GaugeValue, time.Since(begun).Seconds())
}

func (c *leaseCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.activeDesc

	c.scrapesTotalMetric.Describe(ch)
	c.scrapeErrorsTotalMetric.Describe(ch)
	ch <- c.lastScrapeErrorDesc
	ch <- c.lastScrapeTimestampDesc
	ch <- c.lastScrapeDurationDesc
}
