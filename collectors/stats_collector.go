package collectors

import (
	"github.com/DRuggeri/dhcpdleasesreader"
	"github.com/prometheus/client_golang/prometheus"
	"time"
	"sync"
)

type StatCollector struct {
	validDesc   *prometheus.Desc
	expiredDesc *prometheus.Desc
	countDesc   *prometheus.Desc
	modTimeDesc *prometheus.Desc
	info        *dhcpdleasesreader.DhcpdInfo
	mux	*sync.Mutex

	scrapesTotalMetric      prometheus.Counter
	scrapeErrorsTotalMetric prometheus.Counter
	lastScrapeErrorDesc     *prometheus.Desc
	lastScrapeTimestampDesc *prometheus.Desc
	lastScrapeDurationDesc  *prometheus.Desc
}

func NewStatsCollector(namespace string, info *dhcpdleasesreader.DhcpdInfo, mux *sync.Mutex) *StatCollector {
	validDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "stats", "valid"),
		"The number of leases in dhcpd.leases that have not yet expired",
		nil, nil,
	)

	expiredDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "stats", "expired"),
		"The number of leases in dhcpd.leases that have xpired",
		nil, nil,
	)

	countDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "stats", "count"),
		"The number of leases in dhcpd.leases",
		nil, nil,
	)

	modTimeDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "stats", "filetime"),
		"The file timestamp in seconds since epoch of the dhcpd.leases file",
		nil, nil,
	)

	scrapesTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "scrapes_total",
			Help:      "Total number of scrapes",
		},
	)

	scrapeErrorsTotalMetric := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "stats",
			Name:      "scrape_errors_total",
			Help:      "Total number of scrapes errors",
		},
	)

	lastScrapeErrorDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "stats", "last_scrape_error"),
		"Whether the last scrape of stats resulted in an error (1 for error, 0 for success).",
		nil, nil,
	)

	lastScrapeTimestampDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "stats", "last_scrape_timestamp"),
		"Number of seconds since 1970 since last scrape of stat metrics.",
		nil, nil,
	)

	lastScrapeDurationDesc := prometheus.NewDesc(prometheus.BuildFQName(namespace, "stats", "last_scrape_duration_seconds"),
		"Number of seconds since 1970 since last scrape of stat metrics.",
		nil, nil,
	)

	return &StatCollector{
		validDesc:   validDesc,
		expiredDesc: expiredDesc,
		countDesc:   countDesc,
		modTimeDesc: modTimeDesc,
		info:        info,
		mux:	mux,

		scrapesTotalMetric:      scrapesTotalMetric,
		scrapeErrorsTotalMetric: scrapeErrorsTotalMetric,
		lastScrapeErrorDesc:     lastScrapeErrorDesc,
		lastScrapeTimestampDesc: lastScrapeTimestampDesc,
		lastScrapeDurationDesc:  lastScrapeDurationDesc,
	}
}

func (c *StatCollector) Collect(ch chan<- prometheus.Metric) {
	var begun = time.Now()
	err_num := 0

        c.mux.Lock()
        defer c.mux.Unlock()

	/* TODO: Surface read errors through this function */
	c.info.Read()

	c.scrapesTotalMetric.Inc()
	if err_num != 0 {
		c.scrapeErrorsTotalMetric.Inc()
	}

	ch <- prometheus.MustNewConstMetric(c.validDesc, prometheus.GaugeValue, float64(c.info.Valid))
	ch <- prometheus.MustNewConstMetric(c.expiredDesc, prometheus.GaugeValue, float64(c.info.Expired))
	ch <- prometheus.MustNewConstMetric(c.countDesc, prometheus.GaugeValue, float64(len(c.info.Leases)))
	ch <- prometheus.MustNewConstMetric(c.modTimeDesc, prometheus.GaugeValue, float64(c.info.ModTime.Unix()))

	ch <- c.scrapesTotalMetric
	ch <- c.scrapeErrorsTotalMetric
	ch <- prometheus.MustNewConstMetric(c.lastScrapeErrorDesc, prometheus.GaugeValue, float64(err_num))
	ch <- prometheus.MustNewConstMetric(c.lastScrapeTimestampDesc, prometheus.GaugeValue, float64(time.Now().Unix()))
	ch <- prometheus.MustNewConstMetric(c.lastScrapeDurationDesc, prometheus.GaugeValue, time.Since(begun).Seconds())
}

func (c *StatCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.validDesc
	ch <- c.expiredDesc
	ch <- c.countDesc

	c.scrapesTotalMetric.Describe(ch)
	c.scrapeErrorsTotalMetric.Describe(ch)
	ch <- c.lastScrapeErrorDesc
	ch <- c.lastScrapeTimestampDesc
	ch <- c.lastScrapeDurationDesc
}
