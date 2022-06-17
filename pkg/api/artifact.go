package api

type Artifact struct {
	RemotePath string
	LocalPath  string
	URLs       []*SignedURL
}

func RemotePaths(artifacts []*Artifact) []string {
	remotePaths := []string{}
	for _, artifact := range artifacts {
		remotePaths = append(remotePaths, artifact.RemotePath)
	}

	return remotePaths
}
