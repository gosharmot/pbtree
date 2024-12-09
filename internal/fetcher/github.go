package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

var (
	githubRegexp     = regexp.MustCompile(`github.com/(.*?)/(.*?)/(.*[.]proto)`)
	googleapisRegexp = regexp.MustCompile(`google/api/(.*?.proto)`)
	protobufRegexp   = regexp.MustCompile(`google/protobuf/(.*?.proto)`)
)

// GithubFetcher allows fetch proto from GitHub
type GithubFetcher struct {
	httpClient *http.Client
	token      string
}

// NewGithub return GithubFetcher
func NewGithub(token string) GithubFetcher {
	return GithubFetcher{
		httpClient: http.DefaultClient,
		token:      token,
	}
}

func (f GithubFetcher) Fetch(ctx context.Context, module string) (io.ReadCloser, error) {
	const (
		googleapisPath = "repos/googleapis/googleapis/contents/google/api/%s"
		protobufPath   = "repos/protocolbuffers/protobuf/contents/src/google/protobuf/%s"
		rawPath        = "repos/%s/%s/contents/%s"
	)

	switch {
	case googleapisRegexp.MatchString(module):
		module = fmt.Sprintf(googleapisPath, googleapisRegexp.ReplaceAllString(module, "$1"))
	case protobufRegexp.MatchString(module):
		module = fmt.Sprintf(protobufPath, protobufRegexp.ReplaceAllString(module, "$1"))
	case githubRegexp.MatchString(module):
		module = fmt.Sprintf(rawPath, githubRegexp.ReplaceAllString(module, "$1"), githubRegexp.ReplaceAllString(module, "$2"), githubRegexp.ReplaceAllString(module, "$3"))
	default:
		return nil, ErrInappropriateFetcher
	}

	return f.fetch(ctx, module)
}

func (f GithubFetcher) fetch(ctx context.Context, module string) (io.ReadCloser, error) {
	const githubApiUrl = "https://api.github.com/"

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s%s", githubApiUrl, module), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-OBApi-Version", "2022-11-28")
	if f.token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.token))
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed get '%s': status: %d", module, resp.StatusCode)
	}

	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ghResp = struct {
		DownloadUrl string `json:"download_url"`
	}{}

	err = json.Unmarshal(b, &ghResp)
	if err != nil {
		return nil, err
	}

	req, err = http.NewRequestWithContext(ctx, http.MethodGet, ghResp.DownloadUrl, nil)
	if err != nil {
		return nil, err
	}

	protoResp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed get proto: status: %d", resp.StatusCode)
	}

	return protoResp.Body, nil
}
