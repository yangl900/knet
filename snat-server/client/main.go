package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"
)

var counter int64

func main() {
	counter = 0
	for {
		c := atomic.LoadInt64(&counter)
		if c < 6000 {
			atomic.AddInt64(&counter, 1)
			go doRequest()

			if c%100 == 0 {
				fmt.Printf("Current connections: %v\n", c)
			}
		} else {
			fmt.Printf("Current connections: %v Sleep...\n", c)
			time.Sleep(time.Second * 1)
		}
	}
}

func doRequest() {
	c := getClient()
	before := time.Now()
	_, err := c.Get("https://snatserver.azurewebsites.net/hook")
	if err != nil {
		after := time.Now()
		fmt.Printf("Failed to send request: %s Duration: %v\n", err, after.Sub(before).Milliseconds())
	}
	atomic.AddInt64(&counter, -1)
}

func getClient() *http.Client {
	defaultTransport := http.DefaultTransport.(*http.Transport)

	transport := &http.Transport{
		Proxy:                 defaultTransport.Proxy,
		DialContext:           defaultTransport.DialContext,
		MaxIdleConns:          128,
		MaxIdleConnsPerHost:   128,
		IdleConnTimeout:       defaultTransport.IdleConnTimeout,
		TLSHandshakeTimeout:   defaultTransport.TLSHandshakeTimeout,
		ExpectContinueTimeout: defaultTransport.ExpectContinueTimeout,
		TLSClientConfig: &tls.Config{
			MinVersion:    tls.VersionTLS12,
			Renegotiation: tls.RenegotiateNever,
		},
	}

	return &http.Client{
		Transport: transport,
	}
}
