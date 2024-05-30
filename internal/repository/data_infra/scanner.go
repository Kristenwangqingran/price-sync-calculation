package data_infra

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"

	"git.garena.com/shopee/core-server/core-logic/clog"
	"git.garena.com/shopee/core-server/core-logic/cutil"
	"git.garena.com/shopee/data-infra/dataservice-sdk-golang/dataservice"
)

/*
	For the complete guide of scan DI using API: https://sites.google.com/shopee.com/datasuite/data-service/quick-start?authuser=1
	For the dataservice SDK startup Guide: https://confluence.shopee.io/pages/viewpage.action?spaceKey=DI&title=Golang+SDK
*/

//nolint:revive,stylecheck
type Scanner struct {
	clientId       string
	clientSecret   string
	bufferSize     int
	readGoroutines int
	errorHandler   func(error)
}

type Expression struct {
	ParameterName string
	Value         string
}

//nolint:revive,stylecheck
func NewScanner(
	clientId string,
	clientSecret string,
) *Scanner {
	return &Scanner{
		clientId:       clientId,
		clientSecret:   clientSecret,
		bufferSize:     1e4,
		readGoroutines: 3,
	}
}

// WithBufferSize can be used to set the returned channel buffer size
func (s *Scanner) WithBufferSize(size int) *Scanner {
	s.bufferSize = size
	return s
}

// WithReadGoroutines can be used to set the number of read goroutines
func (s *Scanner) WithReadGoroutines(goroutines int) *Scanner {
	s.readGoroutines = goroutines
	return s
}

// WithErrorHandler can be used to set the handler for error
func (s *Scanner) WithErrorHandler(handler func(error)) *Scanner {
	s.errorHandler = handler
	return s
}

// Scan will return a channel
// The default buffer size of the returned channel is 1e4
// If you don't want to use the default, please call WithBufferSize before Scan
//
//nolint:lll,funlen
func (s *Scanner) Scan(ctx context.Context, apiAbbreviation string,
	apiVersion string, expressions []*Expression) chan interface{} {
	clog.Infof(ctx, "Scan with %v/%v: %v", apiAbbreviation, apiVersion, cutil.LazyJSONEncoder(expressions))

	// init Client
	c := dataservice.Client{}
	c.SetEnv(dataservice.LIVE).SetQueryPattern(dataservice.OLAP).SetAppKey(s.clientId).
		SetAppSecret(s.clientSecret).Refresh()

	// init Body
	b := dataservice.Body{}
	// The AddExpression func can be called many times to set multiple expressions
	for _, expression := range expressions {
		b.AddExpression(expression.ParameterName, expression.Value)
	}

	// Create a buffered channel to receive olap results, the queue's default maximum size is 3
	// Because the maximum number of parallel goroutines for fetch olap result is 3
	ch := make(chan []interface{}, 3)

	// Use errgroup to handle errors that occur when calling the Call function
	g := new(errgroup.Group)

	// Create a goroutine to write data in parallel to channel
	// Call function will close the input channel when it errors
	g.Go(func() (err error) {
		return c.Call(apiAbbreviation, apiVersion, b, ch, "")
	})

	result := make(chan interface{}, s.bufferSize)

	// Create multiple goroutines concurrently to read the results in the channel
	// The number of goroutines depends on the actual scenario, default value is 3
	// If the reading speed is slower than the writing speed, you can increase the number of goroutines
	wg := sync.WaitGroup{}
	for i := 0; i < s.readGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for v := range ch {
				for _, i := range v {
					result <- i
				}
			}
		}()
	}
	go func() {
		wg.Wait()
		close(result)
	}()

	go func() {
		// handle errors
		if err := g.Wait(); err != nil {
			clog.Errorf(ctx, "Fail to call api %v/%v, err: %v", apiAbbreviation, apiVersion, err)
			if s.errorHandler != nil {
				s.errorHandler(err)
			}
		}
	}()

	return result
}
