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

// MultiFetcher is a combination of different proto providers
type MultiFetcher struct {
	fetchers []Fetcher
}

// NewMultiFetcher return MultiFetcher
func NewMultiFetcher(fetchers ...Fetcher) Fetcher {
	return MultiFetcher{fetchers: fetchers}
}

func (m MultiFetcher) Fetch(ctx context.Context, module string) (io.ReadCloser, error) {
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
