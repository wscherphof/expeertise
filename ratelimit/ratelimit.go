package ratelimit

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"github.com/wscherphof/expeertise/router"
	"github.com/wscherphof/secure"
	"log"
	"net/http"
	"time"
)

var (
	ErrTooManyRequests = errors.New("429 Too Many Requests")
	ErrInvalidRequest  = errors.New("Invalid Request")
	ErrTokenExpired    = errors.New("Token Expired")
)

const tokenTimeOut = time.Minute

type token struct {
	ip      string
	expires time
	limit   time
}

func NewToken(r *http.Request) (string, error) {
	return secure.NewRequestToken(&token{
		ip:        r.RemoteAddr,
		timestamp: time.Now(),
	})
}

func init() {
	secure.RegisterRequestTokenData(token{})
}

type request struct {
	Timestamp time
	Clear     time
}

type client struct {
	IP       string
	Clear    time
	Requests map[string]request
}

func prev(seconds int, r *http.Request) (err error) {
	// TODO: db stuff
	return
}

func RateLimitHandle(seconds int, handle router.ErrorHandle) router.ErrorHandle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err *router.Error) {
		t := new(token)
		if rate := r.FormValue("_rate"); rate == "" {
			err = router.NewError(ErrInvalidRequest)
			log.Printf("ATTACK: rate limit token missing %#v", r)
		} else if e := secure.RequestToken(rate).Read(t); e != nil {
			err = router.NewError(ErrInvalidRequest)
			log.Printf("ATTACK: rate limit token unreadable %#v", r)
		} else if t.ip != r.RemoteAddr {
			err = router.NewError(ErrInvalidRequest)
			log.Printf("ATTACK: rate limit token invalid address %#v", r)
		} else if t.timestamp.Add(tokenTimeOut).Before(time.Now) {
			err = router.NewError(ErrTokenExpired)
		} else if e := prev(seconds, r); e != nil {
			// TODO
			err = router.NewError(e)
		} else {
			err = handle(w, r, ps)
		}
		return
	}
}
