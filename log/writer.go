package log

import (
	"bytes"
	"sync"
	"unicode/utf8"
)

const writerBufferSize = 64 * 1024

// Writer adapts a byte stream to log messages. Long lines are split into bounded
// chunks without splitting valid UTF-8 characters. Write and Flush are safe for
// concurrent use; the print callback must not call back into this writer.
type Writer struct {
	mu      sync.Mutex
	pending []byte
	print   PrintFunc
	chunked bool
}

// NewWriter creates a writer using print (for example Info or Error) to emit
// messages. print must not be nil. Call Flush after the producer finishes to
// emit the last line even if it has no trailing newline.
func NewWriter(print PrintFunc) *Writer {
	return &Writer{print: print}
}

func (w *Writer) Write(data []byte) (int, error) {

	w.mu.Lock()
	defer w.mu.Unlock()

	size := len(data)
	for len(data) > 0 {
		space := writerBufferSize - len(w.pending)
		newline := bytes.IndexByte(data[:min(len(data), space+1)], '\n')
		if newline >= 0 && newline <= space {
			w.pending = append(w.pending, data[:newline]...)
			line := bytes.TrimSuffix(w.pending, []byte{'\r'})
			if len(line) > 0 || !w.chunked {
				w.print(string(line))
			}
			w.pending = w.pending[:0]
			w.chunked = false
			data = data[newline+1:]
			continue
		}

		if space == 0 {
			w.emitChunk()
			continue
		}

		n := min(space, len(data))
		w.pending = append(w.pending, data[:n]...)
		data = data[n:]
	}
	return size, nil
}

// emitChunk retains an incomplete UTF-8 character and a possible CRLF prefix.
func (w *Writer) emitChunk() {

	end := len(w.pending)
	start := end - 1
	for start > 0 && end-start < utf8.UTFMax && !utf8.RuneStart(w.pending[start]) {
		start--
	}
	if !utf8.FullRune(w.pending[start:]) {
		end = start
	}
	if end > 0 && w.pending[end-1] == '\r' {
		end--
	}

	w.print(string(w.pending[:end]))
	w.chunked = true
	remaining := copy(w.pending, w.pending[end:])
	w.pending = w.pending[:remaining]
}

// Flush emits any remaining text. It does not close the writer or its output.
// Call it after all producers finish writing; repeated calls emit nothing extra.
func (w *Writer) Flush() {

	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.pending) > 0 {
		line := bytes.TrimSuffix(w.pending, []byte{'\r'})
		if len(line) > 0 || !w.chunked {
			w.print(string(line))
		}
		w.pending = w.pending[:0]
	}
	w.chunked = false
}
