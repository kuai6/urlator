package runner

import (
	"context"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_runner_execute(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	t.Run("ok, valid request", func(t *testing.T) {
		c := NewMockHttpClient(ctrl)
		c.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: 100,
		}, nil).Times(1)

		r := &runner{
			client:         c,
			requestTimeout: defaultRequestTimeout,
			maxThreads:     defaultMaxThreads,
		}
		got, err := r.execute(context.Background(), "")
		assert.NoError(t, err)
		assert.Equal(t, uint64(100), got)
	})

	t.Run("ok, negative content length", func(t *testing.T) {
		c := NewMockHttpClient(ctrl)
		c.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: -1,
			Body:          ioutil.NopCloser(strings.NewReader("bar")),
		}, nil).Times(1)

		r := &runner{
			client:         c,
			requestTimeout: defaultRequestTimeout,
			maxThreads:     defaultMaxThreads,
		}
		got, err := r.execute(context.Background(), "")
		assert.NoError(t, err)
		assert.Equal(t, uint64(3), got)
	})

	t.Run("fail, error creating http request", func(t *testing.T) {
		c := NewMockHttpClient(ctrl)
		c.EXPECT().Do(gomock.Any()).Times(0)

		r := &runner{
			client:         c,
			requestTimeout: defaultRequestTimeout,
			maxThreads:     defaultMaxThreads,
		}
		got, err := r.execute(context.Background(), "http://foo\x7f.com/")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "error creating http request"))
		assert.Equal(t, uint64(0), got)
	})

	t.Run("fail, error executing http request", func(t *testing.T) {
		c := NewMockHttpClient(ctrl)
		c.EXPECT().Do(gomock.Any()).Return(nil, errors.New("foo")).Times(1)

		r := &runner{
			client:         c,
			requestTimeout: defaultRequestTimeout,
			maxThreads:     defaultMaxThreads,
		}
		got, err := r.execute(context.Background(), "")
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "error executing http request"))
		assert.Equal(t, uint64(0), got)
	})

	t.Run("fail, http status not OK", func(t *testing.T) {
		c := NewMockHttpClient(ctrl)
		c.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode: http.StatusBadRequest,
		}, nil).Times(1)

		r := &runner{
			client:         c,
			requestTimeout: defaultRequestTimeout,
			maxThreads:     defaultMaxThreads,
		}
		got, err := r.execute(context.Background(), "")
		assert.Error(t, err)
		assert.EqualError(t, err, "http request failed")
		assert.Equal(t, uint64(0), got)
	})

}

func Test_runner_worker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	t.Run("ok, works", func(t *testing.T) {

		c := NewMockHttpClient(ctrl)
		c.EXPECT().Do(gomock.Any()).Return(&http.Response{
			StatusCode:    http.StatusOK,
			ContentLength: 100,
		}, nil).Times(1)

		r := &runner{
			client:         c,
			requestTimeout: defaultRequestTimeout,
			maxThreads:     defaultMaxThreads,
		}

		urls := make(chan string, 1)
		urls <- "https://foo.bar"
		close(urls)

		resps := make(chan Result, 1)
		defer close(resps)
		wg := &sync.WaitGroup{}
		wg.Add(1)
		r.worker(context.Background(), urls, resps, wg)

		result := <-resps

		assert.NoError(t, result.Err)
		assert.Equal(t, result.Size, uint64(100))
		assert.Equal(t, result.Url, "https://foo.bar")
	})

	t.Run("ok,  ctx cancel", func(t *testing.T) {
		r := &runner{
			client:         NewMockHttpClient(ctrl),
			requestTimeout: defaultRequestTimeout,
			maxThreads:     defaultMaxThreads,
		}

		urls := make(chan string, 1)
		close(urls)

		resps := make(chan Result, 1)
		defer close(resps)
		wg := &sync.WaitGroup{}
		wg.Add(1)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		r.worker(ctx, urls, resps, wg)

		assert.Equal(t, len(resps), 0)
	})

}

func Test_runner_Run(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type fields struct {
		client     HttpClient
		maxThreads int
	}
	type args struct {
		ctx  context.Context
		urls []string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Result
		wantErr bool
	}{
		{
			name: "ok, one url",
			fields: func() fields {

				c := NewMockHttpClient(ctrl)
				c.EXPECT().Do(gomock.Any()).Return(&http.Response{
					StatusCode:    http.StatusOK,
					ContentLength: 100,
				}, nil).Times(1)

				return fields{
					client:     c,
					maxThreads: defaultMaxThreads,
				}
			}(),
			args: args{
				ctx:  context.Background(),
				urls: []string{"https://foo.bar"},
			},
			want: []Result{
				{
					Url:  "https://foo.bar",
					Size: 100,
					Err:  nil,
				},
			},
		},
		{
			name: "ok, five urls",
			fields: func() fields {

				c := NewMockHttpClient(ctrl)
				c.EXPECT().Do(gomock.Any()).Return(&http.Response{
					StatusCode:    http.StatusOK,
					ContentLength: 100,
				}, nil).Times(5)

				return fields{
					client:     c,
					maxThreads: defaultMaxThreads,
				}
			}(),
			args: args{
				ctx: context.Background(),
				urls: []string{
					"https://foo1.bar",
					"https://foo2.bar",
					"https://foo3.bar",
					"https://foo4.bar",
					"https://foo5.bar",
				},
			},
			want: []Result{
				{
					Url:  "https://foo1.bar",
					Size: 100,
					Err:  nil,
				},
				{
					Url:  "https://foo2.bar",
					Size: 100,
					Err:  nil,
				},
				{
					Url:  "https://foo3.bar",
					Size: 100,
					Err:  nil,
				},
				{
					Url:  "https://foo4.bar",
					Size: 100,
					Err:  nil,
				},
				{
					Url:  "https://foo5.bar",
					Size: 100,
					Err:  nil,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &runner{
				client:         tt.fields.client,
				requestTimeout: defaultRequestTimeout,
				maxThreads:     tt.fields.maxThreads,
			}
			got, err := r.Run(tt.args.ctx, tt.args.urls)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !assert.ElementsMatch(t, got, tt.want) {
				t.Errorf("Run() got = %v, want %v", got, tt.want)
			}
		})
	}
}
