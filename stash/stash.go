package main

import (
	"flag"
	"time"

	"github.com/olivere/elastic"
	"github.com/tal-tech/go-queue/kq"
	"github.com/tal-tech/go-stash/stash/config"
	"github.com/tal-tech/go-stash/stash/es"
	"github.com/tal-tech/go-stash/stash/filter"
	"github.com/tal-tech/go-stash/stash/handler"
	"github.com/tal-tech/go-zero/core/conf"
	"github.com/tal-tech/go-zero/core/logx"
	"github.com/tal-tech/go-zero/core/proc"
)

const dateFormat = "2006.01.02"

var configFile = flag.String("f", "etc/config.json", "Specify the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	proc.SetTimeoutToForceQuit(c.GracePeriod)

	client, err := elastic.NewClient(
		elastic.SetSniff(false),
		elastic.SetURL(c.Output.ElasticSearch.Hosts...),
	)
	logx.Must(err)

	indexFormat := c.Output.ElasticSearch.DailyIndexPrefix + dateFormat
	var loc *time.Location
	if len(c.Output.ElasticSearch.TimeZone) > 0 {
		loc, err = time.LoadLocation(c.Output.ElasticSearch.TimeZone)
		logx.Must(err)
	} else {
		loc = time.Local
	}
	indexer := es.NewIndex(client, func(t time.Time) string {
		return t.In(loc).Format(indexFormat)
	})

	filters := filter.CreateFilters(c)
	writer, err := es.NewWriter(c.Output.ElasticSearch, indexer)
	logx.Must(err)

	handle := handler.NewHandler(writer)
	handle.AddFilters(filters...)
	handle.AddFilters(filter.AddUriFieldFilter("url", "uri"))
	q := kq.MustNewQueue(c.Input.Kafka, handle)
	q.Start()
}
