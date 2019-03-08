package cmd

import (
	"bufio"
	"bytes"
	"path"
	"testing"
	"time"
)

func TestGCS(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping GCS tests in short mode")
	}
	filename := path.Join("test", "artifact", "x.zip")
	content := []byte("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum.")
	err := writeGCS(filename, bytes.NewReader(content), time.Second*10)
	if err != nil {
		t.Fatalf("failed to write to Google Cloud Storage, err: %s", err)
	}
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	if err = readGCS(writer, filename); err != nil {
		t.Fatalf("failed to read from Google Cloud Storage, err: %s", err)
	}
	writer.Flush()
	if !bytes.Equal(b.Bytes(), content) {
		t.Errorf("downloaded content(%s) doesn't match previously uploaded(%s)", b.String(), string(content))
	}
	if err = delGCS(filename); err != nil {
		t.Fatalf("failed to delete from to Google Cloud Storage, err: %s", err)
	}
}
