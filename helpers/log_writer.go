package helpers

import (
	"bytes"
	"sync"

	"github.com/osixia/container-baseimage/log"
)

// logWriter drains command output without Scanner's maximum line length.
// exec.Cmd waits for its output copy goroutines before Wait returns.
type logWriter struct {
	mu      sync.Mutex
	pending []byte
	print   log.PrintFunc
}

func (w *logWriter) Write(data []byte) (int, error) {

	log.Tracef("logWriter.Write called with bytes: %v", len(data))

	w.mu.Lock()
	defer w.mu.Unlock()

	size := len(data)

	for len(data) > 0 {
		newline := bytes.IndexByte(data, '\n')
		if newline < 0 {
			w.pending = append(w.pending, data...)
			break
		}

		w.pending = append(w.pending, data[:newline]...)
		w.print(string(bytes.TrimSuffix(w.pending, []byte{'\r'})))
		w.pending = nil
		data = data[newline+1:]
	}

	return size, nil

}

func (w *logWriter) Flush() {

	log.Trace("logWriter.Flush called")

	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.pending) > 0 {
		w.print(string(bytes.TrimSuffix(w.pending, []byte{'\r'})))
		w.pending = nil
	}
}
