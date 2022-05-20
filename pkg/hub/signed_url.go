package hub

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	log "github.com/sirupsen/logrus"
)

// SignedURL contains an url and its method type.
type SignedURL struct {
	URL    string `json:"url,omitempty"`
	Method string `json:"method,omitempty"`
}

func (u *SignedURL) GetObject() (string, error) {
	URL, _ := url.Parse(u.URL)

	switch host := URL.Host; {
	case host == "storage.googleapis.com":
		log.Debugf("Parsing GCS URL: %s\n", u.URL)
		return parseGoogleStorageURL(URL)
	case strings.HasSuffix(host, "amazonaws.com"):
		log.Debug("Parsing S3 URL: %s\n", u.URL)
		return parseS3URL(URL)
	case strings.HasPrefix(host, "localhost"):
		log.Debug("Parsing localhost URL: %s\n", u.URL)
		return parseLocalhostURL(URL)
	default:
		log.Warnf("Failed to parse URL '%s' - unrecognized host '%s'\n", u.URL, host)
		return "", fmt.Errorf("unrecognized host %s", host)
	}
}

// GCS URLs follow the format 'https://storage.googleapis.com/<bucket-name>/<path>'
func parseGoogleStorageURL(URL *url.URL) (string, error) {
	re := regexp.MustCompile(`https:\/\/storage\.googleapis\.com\/[a-z0-9\-]+\/([^?]+)\?Expires=`)
	parsed := re.FindStringSubmatch(URL.String())
	if len(parsed) < 2 {
		log.Warn("Failed to parse GCS URL.\n")
		return "", fmt.Errorf("bad URL")
	}

	return parsed[1], nil
}

// S3 URLs follow the format 'https://<bucket-name>.s3.<region>.amazonaws.com/<path>'
// Note: S3 URLs use the project id as a prefix, so we take that into account here as well
func parseS3URL(URL *url.URL) (string, error) {
	re := regexp.MustCompile(`https:\/\/(.+)\.s3\.(.+)\.amazonaws\.com\/[a-z0-9\-]+\/([^?]+)\?`)
	parsed := re.FindStringSubmatch(URL.String())
	if len(parsed) < 4 {
		log.Warn("Failed to parse S3 URL.\n")
		return "", fmt.Errorf("")
	}

	return parsed[3], nil
}

// Localhost URLs are used during tests
func parseLocalhostURL(URL *url.URL) (string, error) {
	return URL.Path, nil
}
