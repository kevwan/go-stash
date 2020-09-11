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
	"github.com/tal-tech/go-zero/core/service"
)

const dateFormat = "2006.01.02"

var configFile = flag.String("f", "etc/config.yaml", "Specify the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	proc.SetTimeoutToForceQuit(c.GracePeriod)

	group := service.NewServiceGroup()
	defer group.Stop()

	for _, processor := range c.Processors {
		client, err := elastic.NewClient(
			elastic.SetSniff(false),
			elastic.SetURL(processor.Output.ElasticSearch.Hosts...),
		)
		logx.Must(err)

		var loc *time.Location
		if len(processor.Output.ElasticSearch.TimeZone) > 0 {
			loc, err = time.LoadLocation(processor.Output.ElasticSearch.TimeZone)
			logx.Must(err)
		} else {
			loc = time.Local
		}
		indexer := es.NewIndex(client, processor.Output.ElasticSearch.Index, loc)
		filters := filter.CreateFilters(processor)
		writer, err := es.NewWriter(processor.Output.ElasticSearch, indexer)
		logx.Must(err)

		handle := handler.NewHandler(writer)
		handle.AddFilters(filters...)
		handle.AddFilters(filter.AddUriFieldFilter("url", "uri"))
		group.Add(kq.MustNewQueue(processor.Input.Kafka, handle))
	}

	group.Start()
}
