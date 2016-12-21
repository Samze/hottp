package main

import (
	"net/http"
	"github.com/samze/hottp"
	"log"
	"os"
	"time"
	"net/url"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Decorator func (c Client) Client

type ClientFunc func(*http.Request) (*http.Response, error)

func (c ClientFunc) Do(req *http.Request) (*http.Response, error) {
	return c(req)
}

func main() {
	var client Client
	client = &http.Client{}

	oneUrl, _ := url.Parse("https://httpbin.org/basic-auth/user/passwd")
	twoUrl, _ := url.Parse("https://httpbin.org/")
	badUrl, _ := url.Parse("https://examplebadthingie.net")

	urls := []*url.URL{oneUrl, twoUrl, badUrl}
	logger := log.New(os.Stdout, "hottp: ", log.Ldate | log.Ltime)

	client = hottp.Decorate(
		client,
		hottp.LoggingDecorator(hottp.StandardHttpLogger(logger)),
		hottp.LoadBalancerDecorator(hottp.RandomStrategy, urls...),
		hottp.AuthorizationDecorator("user", "passwd"),
		hottp.RetryDecorator(5, time.Second * 2, logger),
	)

	req, err := http.NewRequest("GET", "https://httpbin.org/basic-auth/user/passwd", nil)


	if err != nil {
		panic(err)
	}

	res, err := client.Do(req)

	if err != nil {
		panic(err)
	}

	if res.StatusCode != 200 {
		panic(res.StatusCode)
	}

	logger.Printf("%d", res.StatusCode)
}
