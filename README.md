## ABOUT
Simple `GO` scalable chat using:
* `websockets` to receive/delivery the messages from/to clients
* `Redis` that acts as a pub/sub channel
* `Elasticsearch` to store the messages


## PRE-REQS
### Running redis instance
* redis server
~~~
$ podman run -d --name redis -p 6379:6379 redis
~~~

### Running elasticsearch instance
* es single node server (without persistency)
~~~
$ podman run -d --name es -p 9200:9200 -p 9300:9300 -e "discovery.type=single-node" elasticsearch:7.12.0
~~~

### go-redis and es package (only if not building/running as a container)
~~~
go get github.com/go-redis/redis/v8
go get github.com/elastic/go-elasticsearch/v7
~~~

## BUILDING THE APP AS A CONTAINER
~~~
$ git clone https://github.com/git-hyagi/chat_websocket.git
$ cd chat_websocket
$ podman build . -t chat:v0.0.1
~~~

## RUNNING

**BEFORE STARTING THE APP, MAKE SURE TO INSTALL/RUN THE [PRE-REQS](#pre-reqs)**
* make sure that [redis is running](#running-redis-instance)
* make sure that [elasticsearch is running](#running-elasticsearch-instance)

* start the app and frontend as containers
~~~
$ podman run --name chat --rm -d -e REDIS_ADDR=$(hostname -i):6379 -e ES_HOST=http://$(hostname -i):9200 -p 8080:8080 localhost/chat:v0.0.1
$ podman run --name nginx --rm -v .:/usr/share/nginx/html:ro -p 8000:80 -d nginx:1.20-alpine
~~~

## EXTRA

#### Redis
Redis client through a container:
~~~
$ podman run -it --rm --name redis-cli redis redis-cli -h $(hostname -i)
~~~

List channels
~~~
127.0.0.1:6379> PUBSUB channels
~~~

List clients connected
~~~
127.0.0.1:6379> CLIENT list
~~~

Test publishing message
~~~
127.0.0.1:6379> PUBLISH chat 'hello world'
~~~


#### Elasticsearch
Change the number of replicas of es index:
~~~
$ podman exec es curl -s -H 'Content-type: application/json' -XPUT localhost:9200/chat/_settings -d '{"index": {"number_of_replicas": 0}}'
~~~

Query the elasticsearch documents:
~~~
curl -sH 'Content-Type: application/json'  -d '{"size":"1000"}' 'localhost:9200/chat/_search?pretty&filter_path=hits.hits._source.client,hits.hits._source.msg'
~~~
