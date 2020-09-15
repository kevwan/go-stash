English | [简体中文](readme-cn.md)

# go-stash

go-stash is a high performance, free and open source server-side data processing pipeline that ingests data from Kafka, processes it, and then sends it to ElasticSearch. 

go-stash is about 4x throughput more than logstash, and easy to deploy, only one executable file.

![go-stash](doc/flow.png)

## Quick Start

```shell
gostash -f etc/config.yaml
```

config.yaml example as below:

```yaml
Processors:
- Input:
    Kafka:
      Name: gostash
      Brokers:
        - "172.16.186.16:19092"
        - "172.16.186.17:19092"
      Topic: k8slog
      Group: pro
      NumProducers: 16
  Filters:
    - Action: drop
      Conditions:
        - Key: k8s_container_name
          Value: "-rpc"
          Type: contains
        - Key: level
          Value: info
          Type: match
          Op: and
    - Action: remove_field
      Fields:
        - message
        - _source
        - _type
        - _score
        - _id
        - "@version"
        - topic
        - index
        - beat
        - docker_container
    - Action: transfer
      Field: message
      Target: data
  Output:
    ElasticSearch:
      Hosts:
        - "172.16.141.4:9200"
        - "172.16.141.5:9200"
      # {.event} is the value of the json attribute from input
      # {{yyyy-MM-dd}} means date, like 2020-09-09
      Index: {.event}-{{yyyy-MM-dd}}
```
