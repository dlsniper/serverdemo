package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dlsniper/debugger"
)

func fakeTraffic() {
	// Wait for the server to start
	time.Sleep(1 * time.Second)

	pages := []string{"/", "/login", "/logout", "/products", "/product/{productID}", "/basket", "/about"}

	activeConns := make(chan struct{}, 10)

	c := &http.Client{
		Timeout: 10 * time.Second,
	}

	i := int64(0)

	for {
		activeConns <- struct{}{}
		i++

		page := pages[rand.Intn(len(pages))]
		page = strings.Replace(page, "{productID}", "abc-"+strconv.Itoa(int(i)), -1)
		r, err := http.NewRequest(http.MethodGet, "http://localhost:8080"+page, nil)
		if err != nil {
			continue
		}
		r = r.WithContext(context.WithValue(context.Background(), "requestID", i))

		go func(i int64) {
			debugger.Middleware(func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Context().Value("requestID"))
				makeRequest(activeConns, c, r)
			}, func(r *http.Request) []string {
				return []string{
					"request", "automated",
					"page", page,
					"rid", strconv.FormatInt(i, 10),
				}
			})(nil, r)
		}(i)
	}
}

func makeRequest(done chan struct{}, c *http.Client, r *http.Request) {
	defer func() {
		// Unblock the next request from the queue
		<-done
	}()

	resp, err := c.Do(r)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, _ = io.Copy(ioutil.Discard, resp.Body)

	time.Sleep(time.Duration(10+rand.Intn(40)) + time.Millisecond)
}
