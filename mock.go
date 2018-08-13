package datasink

// Mock is a fake Client which allows the received messages to be inspected.
// Messages are stored exactly as they are sent to the client (e.g. still [un]gziped)
type Mock struct {
	Messages map[Stream][]Message
}

// NewMockClient creates a new Mock client.
func NewMockClient() Mock {
	messages := make(map[Stream][]Message)
	return Mock{messages}
}

// Post sends a message to the Mock client.
func (m *Mock) Post(stream Stream, msg Message) error {
	m.Messages[stream] = append(m.Messages[stream], msg)
	return nil
}

// PostGzipped sends a new message to the Mock client.
// The message is stored exactly as it is received.
func (m *Mock) PostGzipped(stream Stream, msg Message) error {
	return m.Post(stream, msg)
}
