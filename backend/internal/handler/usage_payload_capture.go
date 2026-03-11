package handler

import (
	"bytes"

	"github.com/gin-gonic/gin"
)

type usagePayloadCaptureWriter struct {
	gin.ResponseWriter
	body        bytes.Buffer
	maxBytes    int64
	totalBytes  int64
	writeFailed bool
	truncated   bool
}

func newUsagePayloadCaptureWriter(writer gin.ResponseWriter, maxBytes int64) *usagePayloadCaptureWriter {
	return &usagePayloadCaptureWriter{
		ResponseWriter: writer,
		maxBytes:       maxBytes,
	}
}

func (w *usagePayloadCaptureWriter) Write(data []byte) (int, error) {
	n, err := w.ResponseWriter.Write(data)
	if n > 0 {
		w.capture(data[:n])
	}
	if err != nil {
		w.writeFailed = true
	}
	return n, err
}

func (w *usagePayloadCaptureWriter) WriteString(value string) (int, error) {
	n, err := w.ResponseWriter.WriteString(value)
	if n > 0 {
		w.capture([]byte(value[:n]))
	}
	if err != nil {
		w.writeFailed = true
	}
	return n, err
}

func (w *usagePayloadCaptureWriter) capture(data []byte) {
	if w == nil || len(data) == 0 {
		return
	}
	w.totalBytes += int64(len(data))
	if w.maxBytes > 0 {
		remaining := w.maxBytes - int64(w.body.Len())
		if remaining <= 0 {
			w.truncated = true
			return
		}
		if int64(len(data)) > remaining {
			_, _ = w.body.Write(data[:int(remaining)])
			w.truncated = true
			return
		}
	}
	_, _ = w.body.Write(data)
}

func (w *usagePayloadCaptureWriter) Snapshot() ([]byte, string, int64, bool) {
	if w == nil {
		return nil, "", 0, true
	}
	body := append([]byte(nil), w.body.Bytes()...)
	contentType := w.Header().Get("Content-Type")
	return body, contentType, w.totalBytes, !w.writeFailed && !w.truncated
}

func attachUsagePayloadCaptureWriter(c *gin.Context, maxBytes int64) (*usagePayloadCaptureWriter, func()) {
	if c == nil || c.Writer == nil {
		return nil, func() {}
	}
	originalWriter := c.Writer
	captureWriter := newUsagePayloadCaptureWriter(originalWriter, maxBytes)
	c.Writer = captureWriter
	restore := func() {
		if c.Writer == captureWriter {
			c.Writer = originalWriter
		}
	}
	return captureWriter, restore
}
