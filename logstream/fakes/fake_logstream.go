package fakes

import (
	"code.cloudfoundry.org/go-loggregator/rpc/loggregator_v2"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
)

type FakeLogStream struct {
	server *httptest.Server
	lock   sync.Mutex

	validToken string

	lastAuthorization string
	requested         bool

	events       chan *loggregator_v2.Envelope
	closeMessage []byte
	doneChan     chan struct{}
}

func NewFakeLogStream(validToken string) *FakeLogStream {
	return &FakeLogStream{
		validToken:   validToken,
		closeMessage: websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		events:       make(chan *loggregator_v2.Envelope, 100),
		doneChan:     make(chan struct{}),
	}
}

func (f *FakeLogStream) Start() {
	f.server = httptest.NewUnstartedServer(f)
	f.server.Start()
}

func (f *FakeLogStream) Close() {
	close(f.doneChan)
	f.server.Close()
}

func (f *FakeLogStream) URL() string {
	return f.server.URL
}

func (f *FakeLogStream) LastAuthorization() string {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.lastAuthorization
}

func (f *FakeLogStream) Requested() bool {
	f.lock.Lock()
	defer f.lock.Unlock()
	return f.requested
}

func (f *FakeLogStream) AddEvent(event *loggregator_v2.Envelope) {
	f.events <- event
}

func (f *FakeLogStream) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	f.lock.Lock()


	f.lastAuthorization = r.Header.Get("Authorization")
	f.requested = true

	if f.lastAuthorization != f.validToken {
		log.Printf("Bad token passed to firehose: %s", f.lastAuthorization)
		rw.WriteHeader(403)
		r.Body.Close()
		return
	}

	f.lock.Unlock()

	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	m := jsonpb.Marshaler{}
	for {
		select {
		case envelope := <-f.events:
			s, err := m.MarshalToString(&loggregator_v2.EnvelopeBatch{
				Batch: []*loggregator_v2.Envelope{
					envelope,
				},
			})
			if err != nil {
				panic(err)
			}

			_, _ = fmt.Fprintf(rw, "data: %s\n\n", s)
			// Flush the data immediatly instead of buffering it for later.
			flusher.Flush()

		case <-f.doneChan:
			return
		}
	}
}
