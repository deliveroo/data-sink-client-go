# Data Sink Go Client

The `datasink` package implements a Go client for speaking to [`data-sink`](https://github.com/deliveroo/data-sink).

## Usage

To create a `Client` pass the URL of the `data-sink` service and a HTTP client to `NewClient`:

```go
client, err := datasink.NewClient("https://data-sink.example.com", http.DefaultClient)
```

The HTTP client needs to conform to the `datasink.Doer` interface. The default `http.Client` from `net/http` conforms to this interface, as do more advanced clients like [`heimdal`](https://github.com/gojektech/heimdall).

The Data Sink client does not do any kind of retrying or backoff. This should be provided by the HTTP client which is passed in to `NewClient`.

To post messages pass them to the client's `Post`:

```go
stream := datasink.Stream{ID: "important-data"}

bytes, err := json.Marshal(importantData)

err = client.Post(stream, bytes)
```

`Post` will automatically append `\n` on to the end of the data, which makes the data easier to read once Data Sink has archived it to S3.

To post data without this extra `\n`, use `PostGzipped` directly:

```go
stream := datasink.Stream{ID: "important-data"}

bytes, err := json.Marshal(importantData)
compressed, err := datasink.Compress(bytes)

err = client.PostGzipped(stream, compressed)
```

### Retries, Exponential Back-off and Circuit Breaking

The client does not retry requests that fail, or do any kind of circuit breaking. It is expected that the HTTP client (which must implement the `datasink.Doer` interface) provides this functionality, as well as any tracing that may be necessary.

It is very easy to implement retry (with exponential back-off) and circuit breaking using the [`heimdal`](https://github.com/gojektech/heimdall) package:

```go
// Exponential back-off settings:
initalTimeout := 2*time.Millisecond
maxTimeout := 9*time.Millisecond
exponentFactor := 2
maximumJitterInterval := 2*time.Millisecond

backoff := heimdall.NewExponentialBackoff(
    initalTimeout,
    maxTimeout,
    exponentFactor,
    maximumJitterInterval
)
retrier := heimdall.NewRetrier(backoff)

timeout := 1000 * time.Millisecond

client := httpclient.NewClient(
    httpclient.WithTimeout(timeout),

)
client := hystrix.NewClient(
    hystrix.WithTimeout(timeout),

    // Retry with exponential back-off:
    httpclient.WithRetrier(retrier),
    httpclient.WithRetryCount(4),

    // Hystrix-like circuit breaking:
    hystrix.WithCommandName("data-sink"),
    hystrix.WithHystrixTimeout(timeout),
    hystrix.WithMaxConcurrentRequests(100),
    hystrix.WithErrorPercentThreshold(20),
    hystrix.WithSleepWindow(10),
    hystrixWithRequestVolumeThreshold(10),
})

client, err := datasink.NewClient("https://data-sink.example.com", client)

// ...
```
