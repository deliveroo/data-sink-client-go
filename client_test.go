package datasink_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/deliveroo/data-sink-client-go"
	. "github.com/onsi/gomega"
)

type server struct {
	*httptest.Server

	// Information about the last request received by the server for assertions:
	uri  *url.URL
	body *[]byte
}

func newServer(status int) server {
	var (
		uri  url.URL
		body []byte
	)

	handler := func(w http.ResponseWriter, r *http.Request) {
		uri = *r.URL

		var err error
		body, err = ioutil.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("warning: error reading body")
		}

		w.WriteHeader(status)
	}
	s := httptest.NewServer(http.HandlerFunc(handler))

	return server{
		Server: s,
		uri:    &uri,
		body:   &body,
	}
}

func (s *server) hostname() string {
	url, err := url.Parse(s.URL)
	if err != nil {
		log.Fatalf("Could not parse URL %v: %v", url, err)
	}

	return url.Hostname()
}

func decompress(payload []byte) []byte {
	buf := bytes.NewBuffer(payload)
	gz, err := gzip.NewReader(buf)
	if err != nil {
		log.Fatalf("Decompress gzip.NewReader")
	}
	defer gz.Close()

	bytes, err := ioutil.ReadAll(gz)
	if err != nil {
		log.Fatalf("Decompress ioutil.ReadAll")
	}

	return bytes
}
func TestClient(t *testing.T) {
	t.Run("NewClient", func(t *testing.T) {
		ts := newServer(200)

		t.Run("Returns client", func(t *testing.T) {
			g := NewGomegaWithT(t)

			c, err := datasink.NewClient("http://example.com", ts.Client())
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(c).ToNot(BeNil())
		})

		t.Run("Invalid URL", func(t *testing.T) {
			g := NewGomegaWithT(t)

			c, err := datasink.NewClient("", ts.Client())
			g.Expect(err).To(HaveOccurred())
			g.Expect(c).To(BeNil())
		})
	})

	t.Run("PostGzipped", func(t *testing.T) {
		g := NewGomegaWithT(t)

		ts := newServer(200)

		c, err := datasink.NewClient(ts.URL, ts.Client())
		g.Expect(err).ToNot(HaveOccurred())

		t.Run("Without PartitionKey", func(t *testing.T) {
			g := NewGomegaWithT(t)

			stream := datasink.Stream{ID: "some-stream-id1"}

			err = c.PostGzipped(stream, nil)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(ts.uri.Path).To(Equal("/archives/some-stream-id1"))
			g.Expect(ts.uri.Query()).To(BeEmpty())
		})

		t.Run("With PartitionKey", func(t *testing.T) {
			g := NewGomegaWithT(t)

			stream := datasink.Stream{
				ID:           "some-stream-id2",
				PartitionKey: "some-partition-key",
			}

			err = c.PostGzipped(stream, nil)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(ts.uri.Path).To(Equal("/archives/some-stream-id2"))
			g.Expect(ts.uri.Query()).To(HaveKeyWithValue("partition_key", []string{"some-partition-key"}))
			g.Expect(ts.uri.Query()).To(HaveLen(1))
		})

		t.Run("Send body unmodified", func(t *testing.T) {
			g := NewGomegaWithT(t)

			stream := datasink.Stream{ID: "some-stream-id"}
			msg := []byte("some body")

			err = c.PostGzipped(stream, msg)
			g.Expect(err).ToNot(HaveOccurred())

			g.Expect(*ts.body).To(Equal([]byte("some body")))
		})

		t.Run("Server returns an error", func(t *testing.T) {
			g := NewGomegaWithT(t)

			ts := newServer(401)

			c, err := datasink.NewClient(ts.URL, http.DefaultClient)
			g.Expect(err).ToNot(HaveOccurred())

			stream := datasink.Stream{ID: "some-stream"}
			msg := []byte("some body")

			err = c.Post(stream, msg)
			g.Expect(err).To(HaveOccurred())
		})
	})

	t.Run("Post", func(t *testing.T) {
		g := NewGomegaWithT(t)

		ts := newServer(200)

		c, err := datasink.NewClient(ts.URL, http.DefaultClient)
		g.Expect(err).ToNot(HaveOccurred())

		t.Run("Compresses body with new line", func(t *testing.T) {
			g := NewGomegaWithT(t)

			stream := datasink.Stream{ID: "some-stream"}
			msg := []byte("some body")

			err = c.Post(stream, msg)
			g.Expect(err).ToNot(HaveOccurred())

			data := decompress(*ts.body)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(data).To(Equal([]byte("some body\n")))
		})
	})
}
