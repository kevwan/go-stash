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

func toKqConf(c config.KafkaConf) ([]kq.KqConf, error) {
	// Fetch topics based on pattern or explicit list
	topics, err := config.FetchMatchingTopics(c)
	if err != nil {
		return nil, err
	}

	var ret []kq.KqConf
	for _, topic := range topics {
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
			Username:    c.Username,
			Password:    c.Password,
		})
	}

	return ret, nil
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
			elastic.SetBasicAuth(processor.Output.ElasticSearch.Username, processor.Output.ElasticSearch.Password),
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

		kqConfs, err := toKqConf(processor.Input.Kafka)
		logx.Must(err)

		for _, k := range kqConfs {
			// Create a handler with topic information for metadata injection
			handle := handler.NewHandlerWithTopic(writer, indexer, k.Topic)
			handle.AddFilters(filters...)
			handle.AddFilters(filter.AddUriFieldFilter("url", "uri"))
			group.Add(kq.MustNewQueue(k, handle))
		}
	}

	group.Start()
}
