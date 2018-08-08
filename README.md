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

err = datasink.Post(stream, bytes)
```

`Post` will automatically append `\n` on to the end of the data, which makes the data easier to read once Data Sink has archived it to S3.

To post data without this extra `\n`, use `PostGzipped` directly:

```go
stream := datasink.Stream{ID: "important-data"}

bytes, err := json.Marshal(importantData)
compressed, err := datasink.Compress(bytes)

err = datasink.PostGzipped(stream, compressed)
```
