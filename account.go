package main

import (
  "net/http"
  "errors"
  "github.com/julienschmidt/httprouter"
)

func SignUpForm (w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
  T("signup", "", map[string]interface{}{
    "Countries": Countries(),
  })(w, r, ps)
}

type Account struct{
  UID string
  PWD string
  Country string
  Postcode string
  FirstName string
  LastName string
}

const accountTable = "account"

func SignUp (w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
  // uid := r.FormValue("uid")
  pwd1 := r.FormValue("pwd1")
  pwd2 := r.FormValue("pwd2")
  if pwd1 != pwd2 {
    Error(w, r, ps, errors.New("Passwords not equal"))
  }
  // TODO: further validation, unique key on UID, password hashing, ...
  record := Account{
    UID: r.FormValue("uid"),
    PWD: pwd1,
    Country: r.FormValue("country"),
    Postcode: r.FormValue("postcode"),
    FirstName: r.FormValue("firstname"),
    LastName: r.FormValue("lastname"),
  }
  Error(w, r, ps, db.Insert(accountTable, record))
}
