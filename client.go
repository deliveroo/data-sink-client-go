package datasink

import (
	"bytes"
	"compress/gzip"
	"net/http"

	"github.com/pkg/errors"
)

// Stream describes the Data Sink stream which a Message is to be sent to.
type Stream struct {
	ID           string
	PartitionKey string
}

// Message is the payload to be sent to the Data Sink stream.
type Message []byte

// Client is an interface for sending data to Data Sink.
type Client interface {
	// Post sends a Message to Data Sink.
	// A new line is appended to the body and the body gzipped before being sent.
	Post(Stream, Message) error

	// Post sends an already gzipped Message to Data Sink.
	PostGzipped(Stream, Message) error
}

// Doer is a generic HTTP client interface.
// It is implemented by net/http's *http.Client, as well as other more advanced
// clients like heimdall.Client.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type client struct {
	url  string
	http Doer
}

const (
	endpoint = "/archives"
)

// NewClient creates a new Data Sink client using the URL of the server and a
// HTTP client that conforms to the Doer interface.
// If requests should be retried or done via a circuit breaker, this must be
// provided by the HTTP client.
func NewClient(url string, http Doer) (Client, error) {
	if url == "" {
		return nil, errors.New("datasink: blank URL")
	}
	return &client{
		url:  url,
		http: http,
	}, nil
}

func (c *client) Post(stream Stream, msg Message) error {
	// Add a newline before compressing so that the files in S3 are readable.
	msg = append(msg, '\n')

	body, err := Compress(msg)
	if err != nil {
		return err
	}
	return c.PostGzipped(stream, body)
}

func (c *client) PostGzipped(stream Stream, msg Message) error {
	url := c.url + endpoint + "/" + stream.ID
	if stream.PartitionKey != "" {
		url += "?partition_key=" + stream.PartitionKey
	}

	buf := bytes.NewBuffer(msg)

	req, err := http.NewRequest(http.MethodPost, url, buf)
	if err != nil {
		return errors.Wrapf(err, "datasink: creating request")
	}
	req.Header.Set("Content-Encoding", "application/gzip")
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := c.http.Do(req)
	if err != nil {
		return errors.Wrapf(err, "datasink: making request")
	}

	if resp.StatusCode >= 400 {
		return errors.Errorf("datasink: returned non-OK response %d", resp.StatusCode)
	}

	return nil
}

// Compress GZIPs a message payload.
func Compress(body []byte) ([]byte, error) {
	buf := bytes.NewBuffer([]byte{})
	gz := gzip.NewWriter(buf)

	_, err := gz.Write(body)
	if err != nil {
		return nil, errors.Wrapf(err, "datasink: compressing body")
	}

	err = gz.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "datasink: close gzip")
	}

	return buf.Bytes(), nil
}
