package runner

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

//go:generate mockgen -source=runner.go -package=runner -destination http_client_mock.go
type HttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

const defaultRequestTimeout = 5 * time.Second
const defaultMaxThreads = 4

type runner struct {
	client         HttpClient
	requestTimeout time.Duration
	maxThreads     int
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

func WithMaxThreads(n int) Option {
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
	tasksNum := len(urls)
	tasks := make(chan string, tasksNum)
	results := make(chan Result, tasksNum)
	defer close(results)

	wg := &sync.WaitGroup{}
	for i := 0; i < r.maxThreads; i++ {
		wg.Add(1)
		go r.worker(ctx, tasks, results, wg)
	}

	go func() {
		for _, url := range urls {
			tasks <- url
		}
		close(tasks)
	}()

	wg.Wait()

	var res []Result
	for {
		done := false
		select {
		case val := <-results:
			res = append(res, val)
		default:
			done = true
		}
		if done {
			break
		}
	}

	return res, nil
}

func (r *runner) worker(ctx context.Context, tasks <-chan string, results chan<- Result, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case url, ok := <-tasks:
			if !ok {
				return
			}
			res := Result{
				Url: url,
			}
			size, err := r.execute(ctx, url)
			if err != nil {
				res.Err = errors.Wrap(err, "execution error")
				results <- res
				continue
			}
			res.Size = size

			results <- res
		case <-ctx.Done():
			log.Debug().Err(ctx.Err()).Msg("context cancelled")
			return
		}
	}
}

func (r *runner) execute(ctx context.Context, url string) (uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, errors.Wrap(err, "error creating http request")
	}
	req.Header.Add("Accept-Encoding", "gzip")

	resp, err := r.client.Do(req)
	if err != nil {
		return 0, errors.Wrap(err, "error executing http request")
	}
	defer func() {
		if resp.Body != nil {
			if err := resp.Body.Close(); err != nil {
				log.Error().Err(err).Msg("closing request body")
			}
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return 0, errors.New("http request failed")
	}

	var size uint64
	if resp.ContentLength < 0 {
		// try to read body size
		maxmem := 1024 // 1kb stack
		rdsz := func(r io.Reader) uint64 {
			bytes := make([]byte, maxmem)
			var size uint64
			for {
				read, err := r.Read(bytes)
				if err == io.EOF {
					break
				}
				size += uint64(read)
			}
			return size
		}

		size = rdsz(resp.Body)
	} else {
		size = uint64(resp.ContentLength)
	}

	return size, nil
}
