package middleware

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// transactionWriter 暂存写请求的响应，直到事务完成后才发送。
type transactionWriter struct {
	gin.ResponseWriter
	header http.Header
	body   bytes.Buffer
	status int
	size   int
}

func newTransactionWriter(writer gin.ResponseWriter) *transactionWriter {
	return &transactionWriter{ResponseWriter: writer, header: writer.Header().Clone(), status: http.StatusOK, size: -1}
}

func (w *transactionWriter) Header() http.Header { return w.header }
func (w *transactionWriter) Status() int         { return w.status }
func (w *transactionWriter) Size() int           { return w.size }
func (w *transactionWriter) Written() bool       { return w.size >= 0 }
func (w *transactionWriter) WriteHeader(status int) {
	if !w.Written() && status > 0 {
		w.status = status
	}
}
func (w *transactionWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
	}
}
func (w *transactionWriter) Write(data []byte) (int, error) {
	w.WriteHeaderNow()
	n, err := w.body.Write(data)
	w.size += n
	return n, err
}
func (w *transactionWriter) WriteString(data string) (int, error) { return w.Write([]byte(data)) }

// 写事务不支持流式响应，Flush 只固定暂存响应头。
func (w *transactionWriter) Flush() { w.WriteHeaderNow() }
func (w *transactionWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	return nil, nil, errors.New("写事务不支持连接接管")
}
func (w *transactionWriter) Pusher() http.Pusher { return nil }

func (w *transactionWriter) send() error {
	header := w.ResponseWriter.Header()
	for key := range header {
		delete(header, key)
	}
	for key, values := range w.header {
		header[key] = append([]string(nil), values...)
	}
	w.ResponseWriter.WriteHeader(w.status)
	if w.body.Len() > 0 {
		_, err := w.ResponseWriter.Write(w.body.Bytes())
		return err
	}
	w.ResponseWriter.WriteHeaderNow()
	return nil
}
