package main

import (
  "net/http"
  "github.com/julienschmidt/httprouter"
  "github.com/wscherphof/expeertise/secure"
)

func Protected (w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
  account := secure.Authentication(r)
  T("protected", "lang", map[string]interface{}{
    "uid": account.UID,
  })(w, r, ps)
}
