package testsupport

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/semaphoreci/artifact/pkg/api"
)

type StorageMockServer struct {
	Server  *httptest.Server
	Handler http.Handler
	Files   []string
}

func NewStorageMockServer() *StorageMockServer {
	return &StorageMockServer{}
}

func (m *StorageMockServer) Init(files []string) {
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
	m.Files = files
	fmt.Printf("Started storage mock at %s\n", mockServer.URL)
}

func (m *StorageMockServer) handleHEADRequest(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Path[1:]
	if !m.isFile(object) {
		w.WriteHeader(404)
	}
}

func (m *StorageMockServer) handleGETRequest(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Path[1:]
	if m.isFile(object) {
		w.Write([]byte("something"))
	} else {
		w.WriteHeader(404)
	}
}

func (m *StorageMockServer) handlePUTRequest(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Path[1:]
	m.addFile(object)
}

func (m *StorageMockServer) handleDELETERequest(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Path[1:]
	if m.isFile(object) {
		m.removeFile(object)
	} else {
		w.WriteHeader(404)
	}
}

func (m *StorageMockServer) URL() string {
	return m.Server.URL
}

func (m *StorageMockServer) Host() string {
	return m.Server.Listener.Addr().String()
}

func (m *StorageMockServer) PushURLs(paths []string, force bool) ([]*api.SignedURL, error) {
	signedURLs := []*api.SignedURL{}
	for _, path := range paths {
		if force {
			signedURLs = append(signedURLs, &api.SignedURL{
				URL:    fmt.Sprintf("%s/%s", m.URL(), path),
				Method: "HEAD",
			})
		}

		signedURLs = append(signedURLs, &api.SignedURL{
			URL:    fmt.Sprintf("%s/%s", m.URL(), path),
			Method: "PUT",
		})
	}

	return signedURLs, nil
}

func (m *StorageMockServer) PullURLs(paths []string) ([]*api.SignedURL, error) {
	path := paths[0]

	if m.isFile(path) {
		return []*api.SignedURL{
			{URL: fmt.Sprintf("%s/%s", m.URL(), path), Method: "GET"},
		}, nil
	}

	if m.isDir(path) {
		signedURLs := []*api.SignedURL{}
		for _, file := range m.findFilesInDir(path) {
			signedURLs = append(signedURLs, &api.SignedURL{
				URL:    fmt.Sprintf("%s/%s", m.URL(), file),
				Method: "GET",
			})
		}

		return signedURLs, nil
	}

	return nil, fmt.Errorf("%s does not exist", path)
}

func (m *StorageMockServer) YankURLs(paths []string) ([]*api.SignedURL, error) {
	path := paths[0]

	if m.isFile(path) {
		return []*api.SignedURL{
			{URL: fmt.Sprintf("%s/%s", m.URL(), path), Method: "DELETE"},
		}, nil
	}

	if m.isDir(path) {
		signedURLs := []*api.SignedURL{}
		for _, file := range m.findFilesInDir(path) {
			signedURLs = append(signedURLs, &api.SignedURL{
				URL:    fmt.Sprintf("%s/%s", m.URL(), file),
				Method: "DELETE",
			})
		}

		return signedURLs, nil
	}

	return nil, fmt.Errorf("%s does not exist", path)
}

func (m *StorageMockServer) isFile(fileName string) bool {
	for _, file := range m.Files {
		if file == fileName {
			return true
		}
	}

	return false
}

func (m *StorageMockServer) isDir(path string) bool {
	for _, file := range m.Files {
		if strings.HasPrefix(file, path) {
			return true
		}
	}

	return false
}

func (m *StorageMockServer) findFilesInDir(path string) []string {
	files := []string{}
	for _, file := range m.Files {
		if strings.HasPrefix(file, path) {
			files = append(files, file)
		}
	}

	return files
}

func (m *StorageMockServer) addFile(fileName string) {
	m.Files = append(m.Files, fileName)
}

func (m *StorageMockServer) removeFile(fileName string) {
	newFiles := []string{}

	for _, file := range m.Files {
		if file != fileName {
			newFiles = append(newFiles, file)
		}
	}

	m.Files = newFiles
}

func (m *StorageMockServer) Close() {
	m.Server.Close()
}
