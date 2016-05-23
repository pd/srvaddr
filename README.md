# srvaddr

Small command-line utility for querying SRV DNS records and rendering the results into a template
or an easily manipulated JSON structure.

## Installation

The usual:

    $ go get -u github.com/pd/srvaddr
    $ go install github.com/pd/srvaddr
    
You can download binaries from the [Github Releases][]:

    $ curl -o srvaddr https://github.com/pd/srvaddr/releases/download/v0.1/srvaddr_linux_amd64 && \
        chmod +x ./srvaddr && \
        ./srvaddr -h

Alternatively, it's available as a [docker image][]:

    $ docker run philodespotos/srvaddr _xmpp-server._tcp.google.com
    xmpp-server.l.google.com:5269
    alt1.xmpp-server.l.google.com:5269
    alt2.xmpp-server.l.google.com:5269
    alt3.xmpp-server.l.google.com:5269
    alt4.xmpp-server.l.google.com:5269

## Examples

The default template simply lists the hostname(s) and port(s) returned for the given query:

    $ srvaddr _frontend._tcp.tld
    box2.example.com:1234
    box3.example.com:7654
    box1.example.com:9876

The results are displayed in no particular order.

> NB: Nothing I use in practice actually populates the priority or weight values in SRV
> records, but the results should _probably_ be sorted based on those.

To lookup multiple services at once, the default unlabeled output is ambiguous. You can label
each service and use the "env-style" template:

    $ srvaddr ZK=_zookeeper._tcp.service.consul MQ=_rabbit._tcp.service.consul
    ZK_ADDR0=spof.node.dc1.consul:2181
    ZK_HOST0=spof.node.dc1.consul
    ZK_PORT0=2181
    ZK_ADDR1=spof.node.dc1.consul:2181
    ZK_HOST1=spof.node.dc1.consul
    ZK_PORT1=2181
    MQ_ADDR0=spof.node.dc1.consul:5672
    MQ_HOST0=spof.node.dc1.consul
    MQ_PORT0=5672

Under the hood, the output formats are all just defined with [text/template][]. You can use
a custom template with `-t filename` (with `-` representing stdin, as usual):

    $ echo '{{with index .db 0}}database port is {{.Port}}{{end}}' | \
        srvaddr -t - db=_postgres._tcp.internal
    database port is 16273

Go's templating can be unwieldy at times. You can just emit JSON, maybe pipe it into [jq][]:

    $ srvaddr -json zk=_zookeeper._tcp.service.consul | jq .
    {
      "zk": [
        {
          "Host": "deadbeef.node.dc1.consul.",
          "IP": "10.2.2.2",
          "Port": 2181
        },
        {
          "Host": "beadfeed.node.dc1.consul.",
          "IP": "10.1.1.1",
          "Port": 2181
        }
      ]
    }

Not all DNS servers will return the additional `A` records necessary for the IP
to be known; favor using the hostname, which is part of the SRV record itself,
whenever possible.

By default, `srvaddr` uses a nameserver from `resolv.conf`; to use a different nameserver:

    $ srvaddr -ns 127.0.0.1:8600 _kafka._tcp.service.consul

## Development

I'll probably write tests at some point. Setting up the mocks was too distracting early on.
Meanwhile, a fairly low friction way of serving up some SRV records is to run consul locally.
Here's a minimal example config:

~~~json
{
  "server": true,
  "bootstrap_expect": 1,
  "log_level": "debug",
  "services": [
    { "id": "zk1", "name": "zk", "address": "10.1.1.1", "port": 2181 },
    { "id": "zk2", "name": "zk", "address": "10.2.2.2", "port": 2181 },
    { "name": "api", "address": "10.3.3.3", "port": 443 },
    { "name": "mq", "address": "127.0.0.1", "port": 5672 }
  ]
}
~~~

[Github Releases]: https://github.com/pd/srvaddr/releases
[docker image]: https://hub.docker.com/r/philodespotos/srvaddr/
[text/template]: https://godoc.org/pkg/text/template
[jq]: https://stedolan.github.io/jq/
