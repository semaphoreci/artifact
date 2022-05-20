package testsupport

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
)

type HubMockServer struct {
	Server         *httptest.Server
	Handler        http.Handler
	StorageBaseURL string
}

func NewHubMockServer(storageBaseURL string) *HubMockServer {
	return &HubMockServer{StorageBaseURL: storageBaseURL}
}

func (m *HubMockServer) Init() {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/api/v1/artifacts") {
			m.handleRequest(w, r)
		} else {
			w.WriteHeader(404)
		}
	}))

	m.Server = mockServer
}

func (m *HubMockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	request := gcs.GenerateSignedURLsRequest{}
	bytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("[HUB MOCK] Error reading request body: %v\n", err)
		w.WriteHeader(500)
		return
	}

	err = json.Unmarshal(bytes, &request)
	if err != nil {
		fmt.Printf("[HUB MOCK] Error unmarshaling request: %v\n", err)
		w.WriteHeader(500)
		return
	}

	fmt.Printf("[HUB MOCK] Received request: %v\n", request)

	response := &gcs.GenerateSignedURLsResponse{
		Urls:  []*gcs.SignedURL{},
		Error: "",
	}

	data, err := json.Marshal(response)
	if err != nil {
		fmt.Printf("[HUB MOCK] Error marshaling response: %v\n", err)
		w.WriteHeader(500)
		return
	}

	_, _ = w.Write(data)
}

func (m *HubMockServer) generateUrls(request *gcs.GenerateSignedURLsRequest) []*gcs.SignedURL {
	// TODO:

}

func (m *HubMockServer) URL() string {
	return m.Server.URL
}

func (m *HubMockServer) Host() string {
	return m.Server.Listener.Addr().String()
}

func (m *HubMockServer) Close() {
	m.Server.Close()
}
