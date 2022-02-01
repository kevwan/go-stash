package main

import (
	"flag"
	"time"

	"github.com/kevwan/go-stash/stash/config"
	"github.com/kevwan/go-stash/stash/es"
	"github.com/kevwan/go-stash/stash/filter"
	"github.com/kevwan/go-stash/stash/handler"
	"github.com/olivere/elastic/v7"
	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/proc"
	"github.com/zeromicro/go-zero/core/service"
)

var configFile = flag.String("f", "etc/config.yaml", "Specify the config file")

func toKqConf(c config.KafkaConf) []kq.KqConf {
	var ret []kq.KqConf

	for _, topic := range c.Topics {
		ret = append(ret, kq.KqConf{
			ServiceConf: c.ServiceConf,
			Brokers:     c.Brokers,
			Group:       c.Group,
			Topic:       topic,
			Offset:      c.Offset,
			Conns:       c.Conns,
			Consumers:   c.Consumers,
			Processors:  c.Processors,
			MinBytes:    c.MinBytes,
			MaxBytes:    c.MaxBytes,
		})
	}

	return ret
}

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	proc.SetTimeToForceQuit(c.GracePeriod)

	group := service.NewServiceGroup()
	defer group.Stop()

	for _, processor := range c.Clusters {
		client, err := elastic.NewClient(
			elastic.SetSniff(false),
			elastic.SetURL(processor.Output.ElasticSearch.Hosts...),
		)
		logx.Must(err)

		filters := filter.CreateFilters(processor)
		writer, err := es.NewWriter(processor.Output.ElasticSearch)
		logx.Must(err)

		var loc *time.Location
		if len(processor.Output.ElasticSearch.TimeZone) > 0 {
			loc, err = time.LoadLocation(processor.Output.ElasticSearch.TimeZone)
			logx.Must(err)
		} else {
			loc = time.Local
		}
		indexer := es.NewIndex(client, processor.Output.ElasticSearch.Index, loc)
		handle := handler.NewHandler(writer, indexer)
		handle.AddFilters(filters...)
		handle.AddFilters(filter.AddUriFieldFilter("url", "uri"))
		for _, k := range toKqConf(processor.Input.Kafka) {
			group.Add(kq.MustNewQueue(k, handle))
		}
	}

	group.Start()
}
