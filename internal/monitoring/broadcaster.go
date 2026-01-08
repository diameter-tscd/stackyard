package monitoring

import (
	"sync"
)

type LogEntry struct {
	Level     string `json:"level"`
	Message   string `json:"message"`
	Timestamp string `json:"time"`
}

type LogBroadcaster struct {
	clients map[chan []byte]bool
	mu      sync.Mutex
}

func NewLogBroadcaster() *LogBroadcaster {
	return &LogBroadcaster{
		clients: make(map[chan []byte]bool),
	}
}

// Write satisfies the io.Writer interface.
// It assumes the input is a JSON string (from zerolog).
func (b *LogBroadcaster) Write(p []byte) (n int, err error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Parse JSON to validate/format (optional, here we just broadcast raw bytes)
	// But since we want to send SSE events, we'll keep it as bytes.
	// We copy the slice because p is reused.
	msg := make([]byte, len(p))
	copy(msg, p)

	for clientChan := range b.clients {
		select {
		case clientChan <- msg:
		default:
			// If client channel is full, drop the message or disconnect client
			// For simplicity, we drop.
		}
	}
	return len(p), nil
}

func (b *LogBroadcaster) Subscribe() chan []byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan []byte, 100) // Buffer log messages
	b.clients[ch] = true
	return ch
}

func (b *LogBroadcaster) Unsubscribe(ch chan []byte) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
}
