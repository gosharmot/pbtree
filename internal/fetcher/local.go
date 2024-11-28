package fetcher

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"regexp"
)

// LocalFetcher allows fetch local proto
type LocalFetcher struct{ wd string }

// NewLocal return LocalFetcher
func NewLocal(wd string) LocalFetcher { return LocalFetcher{wd: wd} }

func (f LocalFetcher) Fetch(_ context.Context, file string) (io.ReadCloser, error) {
	if !regexp.MustCompile("^api/.*/.*.proto").MatchString(file) {
		return nil, ErrInappropriateFetcher
	}
	return os.Open(filepath.Join(f.wd, file))
}
