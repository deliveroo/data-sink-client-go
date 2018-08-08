package datasink

// Mock is a fake Client which allows the received messages to be inspected.
// Messages are stored exactly as they are sent to the client (e.g. still [un]gziped)
type Mock struct {
	Messages [][]byte
}

// NewMockClient creates a new Mock client.
func NewMockClient() Mock {
	return Mock{}
}

// Post sends a message to the Mock client.
func (m *Mock) Post(_ Stream, msg Message) error {
	m.Messages = append(m.Messages, msg)
	return nil
}

// PostGzipped sends a new message to the Mock client.
// The message is stored exactly as it is received.
func (m *Mock) PostGzipped(stream Stream, msg Message) error {
	return m.Post(stream, msg)
}
