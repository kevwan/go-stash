[English](readme.md) | 简体中文

# go-stash简介

go-stash是一个高效的从Kafka获取，根据配置的规则进行处理，然后发送到ElasticSearch集群的工具。

go-stash有大概logstash 5倍的吞吐性能，并且部署简单，一个可执行文件即可。

![go-stash](https://pro-public.xiaoheiban.cn/icon/84cc2f235035d7f1da6df512d4ba97b7.png)


### 安装

```shell
cd stash && go build stash.go
```

### Quick Start

```shell
./stash -f etc/config.yaml
```

config.yaml示例如下:

```yaml
Clusters:
- Input:
    Kafka:
      Name: go-stash
      Log:
        Mode: file
      Brokers:
      - "172.16.48.41:9092"
      - "172.16.48.42:9092"
      - "172.16.48.43:9092"
      Topic: ngapplog
      Group: stash
      Conns: 3
      Consumers: 10
      Processors: 60
      MinBytes: 1048576
      MaxBytes: 10485760
      Offset: first
  Filters:
  - Action: drop
    Conditions:
      - Key: status
        Value: 503
        Type: contains
      - Key: type
        Value: "app"
        Type: match
        Op: and
  - Action: remove_field
    Fields:
    - message
    - source
    - beat
    - fields
    - input_type
    - offset
    - "@version"
    - _score
    - _type
    - clientip
    - http_host
    - request_time
  Output:
    ElasticSearch:
      Hosts:
      - "http://172.16.188.73:9200"
      - "http://172.16.188.74:9200"
      - "http://172.16.188.75:9200"
      Index: "go-stash-{{yyyy.MM.dd}}"
      MaxChunkBytes: 5242880
      GracePeriod: 10s
      Compress: false
      TimeZone: UTC
```

## 详细说明

### input

```shell
Conns: 3
Consumers: 10
Processors: 60
MinBytes: 1048576
MaxBytes: 10485760
Offset: first
```
#### Conns
  链接kafka的链接数，链接数依据cpu的核数，一般<= CPU的核数；

#### Consumers
  每个连接数打开的线程数，计算规则为Conns * Consumers，不建议超过分片总数，比如topic分片为30，Conns *Consumers <= 30

#### Processors
  处理数据的线程数量，依据CPU的核数，可以适当增加，建议配置：Conns * Consumers * 2 或 Conns * Consumers * 3，例如：60  或 90

#### MinBytes MaxBytes
  每次从kafka获取数据块的区间大小，默认为1M~10M，网络和IO较好的情况下，可以适当调高

#### Offset
  可选last和false，默认为last，表示从头从kafka开始读取数据


### Filters

```shell
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
    - offset
    - prospector
    - source
    - stream
- Action: transfer
  Field: message
  Target: data

```
#### - Action: drop
  - 删除标识：满足此条件的数据，在处理时将被移除，不进入es
  - 按照删除条件，指定key字段及Value的值，Type字段可选contains(包含)或match(匹配)
  - 拼接条件Op: and，也可写or

#### - Action: remove_field
  移除字段标识：需要移除的字段，在下面列出即可

#### - Action: transfer
  转移字段标识：例如可以将message字段，重新定义为data字段


### Output

#### Index
  索引名称，indexname-{{yyyy.MM.dd}}表示年.月.日，也可以用{{yyyy-MM-dd}}，格式自己定义

#### MaxChunkBytes
  每次往ES提交的bulk大小，默认是5M，可依据ES的io情况，适当的调整

#### GracePeriod
  默认为10s，在程序关闭后，在10s内用于处理余下的消费和数据，优雅退出

#### Compress
  数据压缩，压缩会减少传输的数据量，但会增加一定的处理性能，可选值true/false，默认为false

####  TimeZone
  默认值为UTC，世界标准时间





## ES性能写入测试


### 测试环境
- stash服务器：3台 4核 8G
- es服务器： 15台 16核  64G

### 关键配置

```shell
- Input:
      Conns: 3
      Consumers: 10
      Processors: 60
      MinBytes: 1048576
      MaxBytes: 10485760
  Filters:
  - Action: remove_field
    Fields:
    - message
    - source
    - beat
    - fields
    - input_type
    - offset
    - request_time
  Output:
      Index: "nginx_pro-{{yyyy.MM.d}}"
      Compress: false
      MaxChunkBytes: 5242880
      TimeZone: UTC
```

### 写入速度平均在15W/S以上
![go-stash](https://pro-public.xiaoheiban.cn/icon/ee207a1cb094c0b3dcaa91ae75b118b8.png)


### 微信交流群

加群之前有劳给一个star，一个小小的star是作者们回答问题的动力。

如果文档中未能覆盖的任何疑问，欢迎您在群里提出，我们会尽快答复。

您可以在群内提出使用中需要改进的地方，我们会考虑合理性并尽快修改。

如果您发现bug请及时提issue，我们会尽快确认并修改。

添加我的微信：kevwan，请注明go-stash，我拉进go-stash社区群🤝

### --END
