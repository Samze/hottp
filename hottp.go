package hottp

import (
	"net/http"
	"log"
	"encoding/base64"
	"time"
	"net/url"
	"math/rand"
)

type Client interface {
	Do(*http.Request) (*http.Response, error)
}

type Decorator func (c Client) Client

type ClientFunc func(*http.Request) (*http.Response, error)

func (c ClientFunc) Do(req *http.Request) (*http.Response, error) {
	return c(req)
}

func Decorate(c Client, decorators ...Decorator) Client {
	for _, decorator := range decorators {
		c = decorator(c)
	}
	return c
}

func LoggingDecorator(reqLogger HttpLogger) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(req *http.Request) (*http.Response, error) {
			reqLogger(req)
			return c.Do(req)
		})
	}
}

func SetHeaderDecorator(key, value string) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(req *http.Request) (*http.Response, error){
			req.Header.Set(key, value)
			return c.Do(req)
		})
	}
}

func AuthorizationDecorator(username, password string) Decorator {
	return SetHeaderDecorator("Authorization", basicAuthHeader(username, password))

}

func RetryDecorator(attempts int, retryInterval time.Duration, logger *log.Logger) Decorator {
	return func(c Client) Client {
		return ClientFunc(func(req *http.Request) (res *http.Response, err error) {
			try := func(req *http.Request) (*http.Response, error){
				return c.Do(req)
			}

			for attempts > 0 {
				if res, err = try(req); err == nil {
					return res, err
				}
				logger.Printf("retrying.. %+v", err)
				attempts--
				time.Sleep(retryInterval)
			}

			return res, err
		})
	}
}

type BalanceStrategy func(...*url.URL) *url.URL

func LoadBalancerDecorator(strategy BalanceStrategy, urls ...*url.URL) Decorator{
	return func(c Client) Client {
		return ClientFunc(func(req *http.Request) (*http.Response, error) {
			req.URL = strategy(urls...)
			return c.Do(req)
		})
	}
}

func RandomStrategy(urls ...*url.URL) *url.URL {
	idx := rand.Intn(len(urls))
	return urls[idx]
}


type HttpLogger func(req *http.Request)

func StandardHttpLogger(log *log.Logger) HttpLogger {
	return func(req *http.Request) {
		log.Printf("%s, %s", req.Method, req.URL)
	}
}

func VerboseHttpLogger(log *log.Logger) HttpLogger{
	return func(req *http.Request) {
		log.Printf("%s, %s %s %s", req.Method, req.URL, req.Body)
	}
}

func basicAuthHeader(username, password string) string {
	return "Basic " + base64Auth(username, password)
}

func base64Auth(username, password string) string {
	data := []byte(username + ":" + password)
	return base64.StdEncoding.EncodeToString(data)
}
