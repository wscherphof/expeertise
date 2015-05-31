package secure

import (
  "net/http"
  "github.com/wscherphof/secure"
  "github.com/wscherphof/secure/middleware"
  "github.com/wscherphof/expeertise/model/account"
  "github.com/wscherphof/expeertise/util2"
  "github.com/julienschmidt/httprouter"
)


var (
  UpdateAuthentication  = secure.UpdateAuthentication
  AuthenticationHandler = middleware.AuthenticationHandler
)

func SecureHandle (handle util2.ErrorHandle) (util2.ErrorHandle) {
  return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err *util2.Error) {
    if authentication := secure.Authentication(w, r); authentication != nil {
      middleware.SetAuthentication(r, authentication)
      err = handle(w, r, ps)
    } else {
      secure.Challenge(w, r)
    }
    return
  }
}

func IfSecureHandle (authenticated util2.ErrorHandle, unauthenticated util2.ErrorHandle) (util2.ErrorHandle) {
  return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err *util2.Error) {
    if authentication := secure.Authentication(w, r); authentication != nil {
      middleware.SetAuthentication(r, authentication)
      err = authenticated(w, r, ps)
    } else {
      err = unauthenticated(w, r, ps)
    }
    return
  }
}

func Authentication (r *http.Request) (ret *account.Account) {
  if auth := middleware.Authentication(r); auth != nil {
    acc := auth.(account.Account)
    ret = &acc
  }
  return
}

func init () {
  secure.Configure(account.Account{}, &secureDB{}, func(src interface{}) (dst interface{}, valid bool) {
    if src != nil {
      acc := src.(account.Account)
      valid = acc.Refresh()
      dst = acc
    }
    return
  })
}
