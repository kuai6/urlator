package runner

import (
	"context"
	"net/http"
	"time"
)

type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

const defaultRequestTimeout = 5 * time.Second
const defaultMaxThreads = 4

type runner struct {
	client         HttpClient
	requestTimeout time.Duration
	maxThreads     uint
}

type Option func(*runner)

func WithHttpClient(c *http.Client) Option {
	return func(r *runner) {
		r.client = c
	}
}

func WithRequestTimeout(t time.Duration) Option {
	return func(r *runner) {
		r.requestTimeout = t
	}
}

func WithMaxThreads(n uint) Option {
	return func(r *runner) {
		r.maxThreads = n
	}
}

func NewRunner(opts ...Option) *runner {
	r := &runner{
		client:         http.DefaultClient,
		requestTimeout: defaultRequestTimeout,
		maxThreads:     defaultMaxThreads,
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

type Result struct {
	Url  string
	Size uint64
	Err  error
}

func (r *runner) Run(ctx context.Context, urls []string) ([]Result, error) {
	return nil, nil
}

func (r *runner) run(ctx context.Context, url string) (uint64, error) {
	return 0, nil
}
