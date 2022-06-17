package testsupport

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/semaphoreci/artifact/pkg/api"
	"github.com/semaphoreci/artifact/pkg/hub"
)

type HubMockServer struct {
	Server        *httptest.Server
	Handler       http.Handler
	StorageServer *StorageMockServer
}

func NewHubMockServer(storageServer *StorageMockServer) *HubMockServer {
	return &HubMockServer{StorageServer: storageServer}
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
	fmt.Printf("Started hub mock at %s\n", mockServer.URL)
}

func (m *HubMockServer) handleRequest(w http.ResponseWriter, r *http.Request) {
	request := hub.GenerateSignedURLsRequest{}
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

	signedURLs, err := m.generateUrls(request)
	if err != nil {
		fmt.Printf("[HUB MOCK] Error generating signed URLs: %v\n", err)
		w.WriteHeader(500)
		return
	}

	response := &hub.GenerateSignedURLsResponse{
		Urls:  signedURLs,
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

func (m *HubMockServer) generateUrls(request hub.GenerateSignedURLsRequest) ([]*api.SignedURL, error) {
	switch request.Type {
	case hub.GenerateSignedURLsRequestPUSH:
		return m.StorageServer.PushURLs(request.Paths, false)

	case hub.GenerateSignedURLsRequestPUSHFORCE:
		return m.StorageServer.PushURLs(request.Paths, true)

	case hub.GenerateSignedURLsRequestPULL:
		return m.StorageServer.PullURLs(request.Paths)

	case hub.GenerateSignedURLsRequestYANK:
		return m.StorageServer.YankURLs(request.Paths)

	default:
		return nil, fmt.Errorf("not implemented")
	}
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
