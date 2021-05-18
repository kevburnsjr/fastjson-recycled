# fastjson-recycled

This repository implements sync pool recycling for fastjson parser and scanner pools.

Before:
```go
import "github.com/valyala/fastjson"

func main() {
	var pool fastjson.ParserPool

	var p := pool.Get()
	v, _ := p.Parse(`["json"]`)
	pool.Put(p)
}
```
After:
```go
import "github.com/kevburnsjr/fastjson-recycled"

func main() {
	var pool = recycled.NewParserPool(1000)

	var p := pool.Get()
	v, _ := p.Parse(`["json"]`)
	pool.Put(p)
}
```

tl;dr - Periodic sync pool recycling can reduce GC pause disruption by creating more garbage more consistently.

## Why

`sync.Pool` is a powerful tool for reusing buffers to reduce memory allocations in the hot path.

`fastjson` provides `ParserPool` and `ScannerPool` to improve json parsing efficiency.

However, there exist certain behaviors of pools that may be detrimental to an application's performance at scale.

Below we examine heap size growth as it relates to buffer pools with variable length inputs,
illustrate a common failure scenario and propose a solution that improves resiliency at
a negligible cost to performance.

## Heap Size Growth

Assume we have a steady stream of JSON objects coming in through an API.

The size of these objects ranges from 100 bytes to 1 MB.

Their size follows a logarithmic distribution as depicted below:


```
      Meassage Size

     |
  1M |o
     |
100K | o
     |  o
 10K |   o
     |    oo
  1K |      oo
     |        ooo
 100 |           ooooooooooooooooooooooooooooooooooooooooo
     -----------------------------------------------------
     0                                                 100
```

The average message size is < 12 KB, but one message out of every 100 is 1024 KB in size.

If we used a `ParserPool` to process 100 messages perfectly representing this distribution in parallel,
the pool would contain 100 parsers, and the average buffer size across all parsers would be 12 kB.

However, some parsers would have a buffer size of 1024 KB while other buffers would have a buffer size of 100 B.

```
      Heap size per parser after 1 iteration

      |
  1MB |o
      |
100KB | o
      |  o
 10KB |   o
      |    oo
  1KB |      oo
      |        ooo
 100B |           ooooooooooooooooooooooooooooooooooooooooo
      -----------------------------------------------------
      0                                                 100

      TotalHeap Size: 1.16 MB
```

If we ran this again, process 100 messages perfectly representing this distribution in parallel but in reverse order,
the largest message will not use the same parser. It will instead increase the buffer size
of some other parser.

```
      Heap size per parser after 2 iterations

      |
  1MB |o                                                  o
      |
100KB | o                                                o
      |  o                                              o
 10KB |   o                                            o
      |    oo                                        oo
  1KB |      oo                                    oo
      |        ooo                              ooo
 100B |           oooooooooooooooooooooooooooooo
      -----------------------------------------------------
      0                                                 100

      TotalHeap Size: 2.32 MB
```

So even though the workload hasn't changed, the variable size of the messages ensures that the total heap size
will continue to grow until the point where every buffer in the pool is the size of the largest message.

```
      Heap size per parser after 100,000 iterations

      |
  1MB |oooooooooooooooooooooooooooooooooooooooooooooooooooo
      |
100KB |
      |
 10KB |
      |
  1KB |
      |
 100B |
      -----------------------------------------------------
      0                                                 100

      TotalHeap Size: 100 MB
```

These buffers will never decrease in size until the garbage collector is run. Only buffers not in use are eligible
to be discarded.

[![gif](https://miro.medium.com/max/700/1*UnZTtpKV669Ayb90FeqmuQ.gif)](https://medium.com/swlh/go-the-idea-behind-sync-pool-32da5089df72)

So we have a few problems

1) Buffers extend to their maximum size over time
2) Buffers aren't guaranteed to be recycled
3) There is no limit to the number of buffers available to the pool

## Scalability

As load increases, the heap size of the pool increases as a function of

```
concurrency * maximum message size
```

So if the maximum message size is 1 MB, the response time is 100ms and we're serving 1,000 req/s...

```
(1,000 req/s * 0.1 res/s) * 1 MB = 100 MB
```

If we receive a 10x traffic spike to 10krps, that number increases.

```
(10,000 req/s * 0.1 res/s) * 1 MB = 1 GB
```

In the real world this is typically much worse since increased load often results in increased latency (100ms -> 200ms)

```
(10,000 req/s * 0.2 res/s) * 1 MB = 2 GB
```

If you're on a small internet facing API server, 2 GB might be all the ram you have. Now the GC is going to decide
that you're using too much ram and it's going to kick on at the worst possible time starving you for CPU, doubling
your response latency (200ms -> 400ms)

```
(10,000 req/s * 0.4 res/s) * 1 MB = 4 GB
```

Which makes your app request more parsers from the pool, which is more garbage that needs to be collected.

## Resliency

The scenario above illustrates a negative feedback loop. The system is unlikely to recover on its own without some
external intervention. Typically this results in high latency and/or reduced availability for some period of time
until autoscaling kicks in to provision additional infrasctructure to spread the load.

We're not going to go into the possible systemic upstream and downstream effects of increased API latency that could
trigger other systemic negative feedback loops. Just recognize that this is a scenario we'd like to avoid for any
highly available service we are responsible for operating.

## Recycling

This
