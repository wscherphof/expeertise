package ratelimit

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/wscherphof/essix/template"
	db "github.com/wscherphof/rethinkdb"
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
	table         = "ratelimit"
	index         = "Clear"
	pk            = "IP"
	clearInterval = time.Hour
)

func init() {
	if _, err := db.TableCreatePK(table, pk); err == nil {
		log.Println("INFO: table created:", table)
		if _, err := db.IndexCreate(table, index); err != nil {
			log.Println("ERROR: failed to create index:", table, index, err)
		} else {
			log.Println("INFO: index created:", table, index)
		}
	}
	go func() {
		for {
			time.Sleep(clearInterval)
			limit := time.Now().Unix()
			if resp, err := db.DeleteTerm(db.Between(table, index, nil, limit, true, true)); err != nil {
				log.Printf("WARNING: rate limit clearing failed: %v", err)
			} else {
				log.Printf("INFO: %v rate limit records cleared", resp.Deleted)
			}
		}
	}()
	secure.RegisterRequestTokenData(token{})
}

type path string

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

type requests map[path]time.Time

type client struct {
	IP       string
	Clear    int64
	Requests requests
}

func getClient(ip string) (c *client) {
	c = new(client)
	if err := db.Get(table, ip, c); err != nil {
		if err == db.ErrEmptyResult {
			c.IP = ip
			c.Requests = make(requests)
		} else {
			log.Printf("WARNING: error getting from table %v: %v", table, err)
		}
	}
	return
}

func (c *client) save() (err error) {
	_, err = db.InsertUpdate(table, c)
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
			handle(w, r, ps)
			c.Requests[p] = time.Now()
			clear := time.Now().Add(window).Unix()
			if clear > c.Clear {
				c.Clear = clear
			}
			if e := c.save(); e != nil {
				log.Printf("WARNING: error saving to table %v: %v", table, e)
			}
		}
		return
	}
}
