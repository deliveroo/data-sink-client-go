package datasink_test

import (
	"testing"

	"github.com/deliveroo/data-sink-client-go"

	. "github.com/onsi/gomega"
)

func TestMock(t *testing.T) {
	t.Run("NewMock", func(t *testing.T) {
		g := NewGomegaWithT(t)

		m := datasink.NewMockClient()
		g.Expect(m).ToNot(BeNil())

		// Implements interface
		var _ datasink.Client = &m
	})

	t.Run("Post", func(t *testing.T) {
		g := NewGomegaWithT(t)

		m := datasink.NewMockClient()
		stream := datasink.Stream{
			ID:           "some-stream-id",
			PartitionKey: "some-partition-key",
		}
		msg := []byte("some body")

		err := m.Post(stream, msg)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(m.Messages).To(HaveKeyWithValue(stream, []datasink.Message{msg}))
		g.Expect(m.Messages).To(HaveLen(1))
	})

	t.Run("PostGzipped", func(t *testing.T) {
		g := NewGomegaWithT(t)

		m := datasink.NewMockClient()

		body := []byte("some body")
		gzipped, _ := datasink.Compress(body)
		stream := datasink.Stream{
			ID:           "some-stream-id",
			PartitionKey: "some-partition-key",
		}

		err := m.PostGzipped(stream, gzipped)
		g.Expect(err).ToNot(HaveOccurred())
		g.Expect(m.Messages).To(HaveKeyWithValue(stream, []datasink.Message{gzipped}))
		g.Expect(m.Messages).To(HaveLen(1))
	})
}
