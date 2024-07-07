package stat

import (
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
)

var Client statsd.Statter

func init {
	StatsdClient, err = statsd.NewClient(config.StatsdHost, config.StatsdPrefix)
	if err != nil {
		panic(err)
	}
}

func (n Notification) DurationMetric(d time.Duration) {
	Client.TimingDuration(
		"notifier.duration", d, 1.0,
		statsd.Tag{"host", Hostname},
		statsd.Tag{"name", n.Notifier.Name},
	)

	StatsdClient.Gauge(
		"notifier.timeoutquota.percent", int64(float64(d)/float64(n.Notifier.timeout())*100), 1.0,
		statsd.Tag{"host", Hostname},
		statsd.Tag{"name", n.Notifier.Name},
	)
}
