package routes

import "net/http"

type (
	Response struct {
		StatusCode int
		Header     http.Header
	}
	responseWriter struct {
		http.ResponseWriter
		resp Response
	}
)

func (w *responseWriter) Write(buf []byte) (int, error) {
	n, e := w.ResponseWriter.Write(buf)
	if w.resp.StatusCode == 0 {
		w.resp.StatusCode = http.StatusOK
	}
	return n, e
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.resp.StatusCode = statusCode
}

func (w *responseWriter) CloseNotify() <-chan bool {
	if closeNotifier, ok := w.ResponseWriter.(http.CloseNotifier); ok {
		return closeNotifier.CloseNotify()
	}
	return nil
}

func (w *responseWriter) Flush() {
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func wrapResponseWriter(w http.ResponseWriter) (http.ResponseWriter, *Response) {
	rw := responseWriter{
		ResponseWriter: w,
		resp: Response{
			Header: w.Header(),
		},
	}
	return &rw, &rw.resp
}
