package cli

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/v5/statsd"
)

var StatsdClient statsd.Statter = (*statsd.Client)(nil)

func init() {
	config := &statsd.ClientConfig{
		Address: statsdHost(),
		Prefix:  statsdPrefix(),
	}

	var err error
	StatsdClient, err = statsd.NewClientWithConfig(config)
	if err != nil {
		log.Fatal(err)
	}
}

func StatsdInc(metric string, n int64) {
	StatsdClient.Inc(metric, n, 1.0, statsdTags()...)
}

func StatsdDuration(d time.Duration) {
	if Debug {
		fmt.Printf("%s.duration: %v\n", statsdPrefix(), d)
	}
	StatsdClient.TimingDuration("duration", d, 1.0, statsdTags()...)
}

func HasStatsd() bool {
	return false
}

func statsdHost() string {
	return os.Getenv("STATSD_HOST")
}

func statsdPrefix() string {
	return os.Getenv("STATSD_PREFIX")
}

func statsdTags() []statsd.Tag {
	if os.Getenv("STATSD_TAGS") == "" {
		return nil
	}

	tagStrings := strings.Split(
		strings.TrimPrefix(os.Getenv("STATSD_TAGS"), "#"),
		",",
	)

	var tags [](statsd.Tag)

	for _, t := range tagStrings {
		ta := strings.Split(t, ":")
		if len(ta) != 2 {
			Error("STATSD_TAGS contains a tag without a key and value")
		}
		tags = append(tags, statsd.Tag{ta[0], ta[1]})
	}

	return tags
}
