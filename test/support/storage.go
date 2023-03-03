// #nosec
package testsupport

import (
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"

	"github.com/semaphoreci/artifact/pkg/api"
)

type StorageMockServer struct {
	Server           *httptest.Server
	Handler          http.Handler
	StorageDirectory string
	MaxFailures      int
	RequestCount     int
}

type FileMock struct {
	Name     string
	Contents string
}

func NewStorageMockServer() (*StorageMockServer, error) {
	tmpStorageDir, err := ioutil.TempDir("", "tmp-storage-*")
	if err != nil {
		return nil, err
	}

	return &StorageMockServer{StorageDirectory: tmpStorageDir}, nil
}

func (m *StorageMockServer) SetMaxFailures(maxFailures int) {
	m.MaxFailures = maxFailures
}

func (m *StorageMockServer) Init(files []FileMock) error {
	err := m.createInitialFiles(files)
	if err != nil {
		return err
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		m.RequestCount += 1

		if m.RequestCount <= m.MaxFailures {
			w.WriteHeader(503)
			w.Write([]byte("temporarily unavailable"))
			return
		}

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
	fmt.Printf("Started storage mock at %s\n", mockServer.URL)
	return nil
}

func (m *StorageMockServer) createInitialFiles(files []FileMock) error {
	for _, file := range files {
		parentDir := m.filePath(filepath.Dir(file.Name))
		err := os.MkdirAll(parentDir, 0755)
		if err != nil {
			return err
		}

		err = ioutil.WriteFile(m.filePath(file.Name), []byte(file.Contents), 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *StorageMockServer) handleHEADRequest(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Path[1:]

	if !m.IsFile(object) {
		w.WriteHeader(404)
	}
}

func (m *StorageMockServer) handleGETRequest(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Path[1:]

	if m.IsFile(object) {
		contents, err := ioutil.ReadFile(m.filePath(object))
		if err != nil {
			w.WriteHeader(500)
			return
		}

		w.Write(contents)
	} else {
		w.WriteHeader(404)
	}
}

func (m *StorageMockServer) handlePUTRequest(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Path[1:]
	err := m.addFile(object, r.Body)
	if err != nil {
		fmt.Printf("Error writing to file: %v\n", err)
		w.WriteHeader(500)
	}
}

func (m *StorageMockServer) handleDELETERequest(w http.ResponseWriter, r *http.Request) {
	object := r.URL.Path[1:]

	if m.IsFile(object) {
		err := m.removeFile(object)
		if err != nil {
			w.WriteHeader(500)
		}

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
		if !force {
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

	if m.IsFile(path) {
		return []*api.SignedURL{
			{URL: fmt.Sprintf("%s/%s", m.URL(), path), Method: "GET"},
		}, nil
	}

	if m.IsDir(path) {
		signedURLs := []*api.SignedURL{}
		files, err := m.findFilesInDir(path)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
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

	if m.IsFile(path) {
		return []*api.SignedURL{
			{URL: fmt.Sprintf("%s/%s", m.URL(), path), Method: "DELETE"},
		}, nil
	}

	if m.IsDir(path) {
		signedURLs := []*api.SignedURL{}
		files, err := m.findFilesInDir(path)
		if err != nil {
			return nil, err
		}

		for _, file := range files {
			signedURLs = append(signedURLs, &api.SignedURL{
				URL:    fmt.Sprintf("%s/%s", m.URL(), file),
				Method: "DELETE",
			})
		}

		return signedURLs, nil
	}

	return nil, fmt.Errorf("%s does not exist", path)
}

func (m *StorageMockServer) filePath(fileName string) string {
	return fmt.Sprintf("%s/%s", m.StorageDirectory, fileName)
}

func (m *StorageMockServer) IsFile(fileName string) bool {
	fmt.Printf("[STORAGE MOCK] Checking if file %s exists...\n", fileName)

	fileInfo, err := os.Stat(m.filePath(fileName))
	if err != nil {
		return false
	}

	return !fileInfo.IsDir()
}

func (m *StorageMockServer) IsDir(path string) bool {
	fmt.Printf("[STORAGE MOCK] Checking if directory %s exists...\n", path)

	fileInfo, err := os.Stat(m.filePath(path))
	if err != nil {
		return false
	}

	return fileInfo.IsDir()
}

func (m *StorageMockServer) findFilesInDir(path string) ([]string, error) {
	files := []string{}

	err := filepath.WalkDir(fmt.Sprintf("%s/%s", m.StorageDirectory, path), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relativePath := filepath.ToSlash(path)[len(m.StorageDirectory)+1:]

		if d.IsDir() {
			return nil
		}

		files = append(files, relativePath)
		return nil
	})

	return files, err
}

func (m *StorageMockServer) addFile(fileName string, reader io.ReadCloser) error {
	filePath := m.filePath(fileName)
	err := os.MkdirAll(filepath.Dir(filePath), 0755)
	if err != nil {
		return err
	}

	newFile, err := os.Create(filePath)
	if err != nil {
		return err
	}

	defer newFile.Close()

	if _, err := io.Copy(newFile, reader); err != nil {
		return err
	}

	return nil
}

func (m *StorageMockServer) removeFile(fileName string) error {
	filePath := m.filePath(fileName)
	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return m.removeEmptyParentsRecursively(filePath)
}

func (m *StorageMockServer) removeEmptyParentsRecursively(filePath string) error {
	currentPath := filePath

	for {
		parentPath := filepath.Dir(currentPath)
		files, err := ioutil.ReadDir(parentPath)
		if err != nil {
			return err
		}

		// Parent is not empty, so we stop
		if len(files) > 0 {
			break
		}

		err = os.Remove(parentPath)
		if err != nil {
			return err
		}

		currentPath = parentPath
	}

	return nil
}

func (m *StorageMockServer) Close() {
	if m.Server != nil {
		m.Server.Close()
	}

	m.cleanup()
}

func (m *StorageMockServer) cleanup() {
	err := os.RemoveAll(m.StorageDirectory)
	if err != nil {
		fmt.Printf("Error cleaning temporary storage directory: %v\n", err)
	}
}
