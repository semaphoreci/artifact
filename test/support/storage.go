package testsupport

import (
	"net/http"
	"net/http/httptest"
)

type StorageMockServer struct {
	Server  *httptest.Server
	Handler http.Handler
	Paths   []string
}

func NewStorageMockServer() *StorageMockServer {
	return &StorageMockServer{}
}

func (m *StorageMockServer) Init(initialPaths []string) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "HEAD":
			m.handleHEADRequest(w, r)
		case "GET":
			m.handleGETRequest(w, r)
		case "PUT":
			m.handlePUTRequest(w, r)
		case "DELETE":
			m.handleDELETERequest(w, r)
		default:
			w.WriteHeader(503)
		}
	}))

	m.Server = mockServer
	m.Paths = initialPaths
}

func (m *StorageMockServer) handleHEADRequest(w http.ResponseWriter, r *http.Request) {

}

func (m *StorageMockServer) handleGETRequest(w http.ResponseWriter, r *http.Request) {

}

func (m *StorageMockServer) handlePUTRequest(w http.ResponseWriter, r *http.Request) {

}

func (m *StorageMockServer) handleDELETERequest(w http.ResponseWriter, r *http.Request) {

}

func (m *StorageMockServer) URL() string {
	return m.Server.URL
}

func (m *StorageMockServer) Host() string {
	return m.Server.Listener.Addr().String()
}

func (m *StorageMockServer) Close() {
	m.Server.Close()
}
