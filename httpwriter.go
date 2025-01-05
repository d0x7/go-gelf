package gelf

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

type HTTPWriter struct {
	GelfWriter
	mu             sync.Mutex
	MaxReconnect   int
	ReconnectDelay time.Duration
}

func NewHTTPWriter(proto, addr string) (*HTTPWriter, error) {
	var err error
	w := new(HTTPWriter)
	w.MaxReconnect = DefaultMaxReconnect
	w.ReconnectDelay = DefaultReconnectDelay
	w.proto = "http"
	w.addr = fmt.Sprintf("%s://%s/gelf", proto, addr)

	if w.hostname, err = os.Hostname(); err != nil {
		return nil, err
	}

	return w, nil
}

// WriteMessage sends the specified message to the GELF server
// specified in the call to New().  It assumes all the fields are
// filled out appropriately.  In general, clients will want to use
// Write, rather than WriteMessage.
func (w *HTTPWriter) WriteMessage(m *Message) (err error) {
	buf := newBuffer()
	defer bufPool.Put(buf)
	messageBytes, err := m.toBytes(buf)
	if err != nil {
		return err
	}

	messageBytes = append(messageBytes, 0)

	n, err := w.writeToHTTPWithReconnectAttempts(messageBytes)
	if err != nil {
		return err
	}
	if n != len(messageBytes) {
		return fmt.Errorf("bad write (%d/%d)", n, len(messageBytes))
	}

	return nil
}

func (w *HTTPWriter) Write(p []byte) (n int, err error) {
	file, line := getCallerIgnoringLogMulti(1)

	m := constructMessage(p, w.hostname, w.Facility, file, line)

	if err = w.WriteMessage(m); err != nil {
		return 0, err
	}

	return len(p), nil
}

func (w *HTTPWriter) writeToHTTPWithReconnectAttempts(zBytes []byte) (n int, err error) {
	var i int

	// HTTP-POST-Request
	req, err := http.NewRequest("POST", w.addr, bytes.NewReader(zBytes))
	if err != nil {
		return 0, fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w.mu.Lock()
	for i = 0; i <= w.MaxReconnect; i++ {
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			// Set an error if the request failed so we can retry
			err = fmt.Errorf("failed to send log messages: %v", err)
		} else if resp.StatusCode != http.StatusAccepted {
			time.Sleep(w.ReconnectDelay * time.Second)
			// Return an error if the status code is not OK
			return 0, fmt.Errorf("graylog responded with non-ok status code %d", resp.StatusCode)
		} else {
			n = len(zBytes)
			break
		}
	}
	w.mu.Unlock()

	if i > w.MaxReconnect {
		return 0, fmt.Errorf("maximum reconnection attempts was reached; giving up")
	}
	return n, nil
}
