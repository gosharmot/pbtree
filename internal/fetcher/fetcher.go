package fetcher

import (
	"context"
	"errors"
	"io"
)

// ErrInappropriateFetcher fetcher is not suitable for the file
var ErrInappropriateFetcher = errors.New("inappropriate fetcher")

// Fetcher is the interface that wraps fetching proto files
type Fetcher interface {
	Fetch(ctx context.Context, module string) (io.ReadCloser, error)
}

// CompoundFetcher is a combination of different proto providers
type CompoundFetcher struct {
	fetchers []Fetcher
}

// NewCompoundFetcher return CompoundFetcher
func NewCompoundFetcher(fetchers ...Fetcher) Fetcher {
	return CompoundFetcher{fetchers: fetchers}
}

func (m CompoundFetcher) Fetch(ctx context.Context, module string) (io.ReadCloser, error) {
	for _, fetcher := range m.fetchers {
		fetched, err := fetcher.Fetch(ctx, module)
		if err != nil {
			if errors.Is(err, ErrInappropriateFetcher) {
				continue
			}
			return nil, err
		}
		return fetched, nil
	}
	return nil, ErrInappropriateFetcher
}
