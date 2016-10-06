package ratelimit

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/wscherphof/essix/entity"
	"github.com/wscherphof/essix/template"
	"github.com/wscherphof/secure"
	"log"
	"net/http"
	"strings"
	"time"
)

var (
	ErrTooManyRequests = errors.New("ErrTooManyRequests")
	ErrInvalidRequest  = errors.New("ErrInvalidRequest")
)

const (
	clearInterval = time.Hour
)

type path string

type requests map[path]time.Time

type client struct {
	*entity.Base
	Clear    int64
	Requests requests
}

func init() {
	entity.Register(&client{}).Index("Clear")
	secure.RegisterRequestTokenData(token{})
	go func() {
		for {
			time.Sleep(clearInterval)
			limit := time.Now().Unix()
			index := entity.Index(&client{}, "Clear")
			selection := index.Between(nil, limit, true, true)
			if deleted, err := selection.Delete(); err != nil {
				log.Printf("WARNING: rate limit clearing failed: %v", err)
			} else {
				log.Printf("INFO: %v rate limit records cleared", deleted)
			}
		}
	}()
}

type token struct {
	IP        string
	Path      path
	Timestamp time.Time
}

func ip(r *http.Request) string {
	return strings.Split(r.RemoteAddr, ":")[0]
}

func NewToken(r *http.Request) (string, error) {
	return secure.NewRequestToken(&token{
		IP:        ip(r),
		Path:      path(r.URL.Path),
		Timestamp: time.Now(),
	})
}

func getClient(ip string) (c *client) {
	c = &client{Base: &entity.Base{ID: ip}}
	if err := c.Read(c); err != nil {
		if err == entity.ErrEmptyResult {
			c.ID = ip
			c.Requests = make(requests)
		} else {
			log.Printf("WARNING: error reading from ratelimit table: %v", err)
		}
	}
	return
}

func Handle(handle httprouter.Handle, seconds int) httprouter.Handle {
	window := time.Duration(seconds) * time.Second
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		t, ip, p := new(token), ip(r), path(r.URL.Path)
		if rate := r.FormValue("_rate"); rate == "" {
			template.Error(w, r, ErrInvalidRequest, true)
			log.Printf("SUSPICIOUS: rate limit token missing %v %v", ip, p)
		} else if e := secure.RequestToken(rate).Read(t); e != nil {
			template.Error(w, r, ErrInvalidRequest, true)
			log.Printf("SUSPICIOUS: rate limit token unreadable %v %v", ip, p)
		} else if t.IP != ip {
			template.Error(w, r, ErrInvalidRequest, true)
			log.Printf("SUSPICIOUS: rate limit token invalid address: %v, expected %v %v", t.IP, ip, p)
		} else if t.Path != p {
			template.Error(w, r, ErrInvalidRequest, true)
			log.Printf("SUSPICIOUS: rate limit token invalid path: %v, token path %v, expected %v", ip, t.Path, p)
		} else if c := getClient(ip); c.Requests[p].After(t.Timestamp) {
			template.Error(w, r, ErrInvalidRequest, true)
			log.Printf("SUSPICIOUS: rate limit token reuse: %v %v, token %v, previous request %v", ip, p, t.Timestamp, c.Requests[p])
		} else if c.Requests[p].After(time.Now().Add(-window)) {
			template.DataError(w, r, ErrTooManyRequests, map[string]interface{}{
				"Window": window,
			}, "ratelimit", "toomanyrequests")
		} else {
			c.Requests[p] = time.Now()
			if clear := time.Now().Add(window).Unix(); clear > c.Clear {
				c.Clear = clear
			}
			if e := c.Update(c); e != nil {
				log.Printf("WARNING: error updating ratelimit record: %v", e)
			}
			handle(w, r, ps)
		}
		return
	}
}
