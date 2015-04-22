package account

import (
  "github.com/wscherphof/expeertise/db"
  "github.com/wscherphof/expeertise/util"
  "errors"
  "time"
  "strings"
  "golang.org/x/crypto/bcrypt"
  "log"
)

var (
  ErrInvalidCredentials = errors.New("Unknown email address or incorrect password or activation code")
  ErrPasswordEmpty = errors.New("Password empty")
  ErrPasswordsNotEqual = errors.New("Passwords not equal")
  ErrEmailTaken = errors.New("Email address taken")
  ErrNotActivated = errors.New("Account hasn't been activated yet")
)

const ACCOUNT_TABLE = "account"

func Init () {
  DefineMessages()
  if cursor, _ := db.TableCreatePK(ACCOUNT_TABLE, "UID"); cursor != nil {
    log.Println("INFO: table created:", ACCOUNT_TABLE)
  }
}

type password struct {
 Created time.Time
 Value []byte
}

func newPassword (pwd1, pwd2 string) (pwd *password, err error) {
  if pwd1 == "" {
    err = ErrPasswordEmpty
  } else if pwd1 != pwd2 {
    err = ErrPasswordsNotEqual
  } else if hash, e := bcrypt.GenerateFromPassword([]byte(pwd1), bcrypt.DefaultCost); err != nil {
    err = e
  } else {
    pwd = &password{
      Created: time.Now(),
      Value: hash,
    }
  }
  return
}

type Account struct {
  dirty bool
  Created time.Time
  UID string
  PWD *password
  Country string
  Postcode string
  FirstName string
  LastName string
  ActivationCode string
}

func (a *Account) FullName () (name string) {
  name = ""
  if len(a.FirstName) > 0 {
    name = a.FirstName
  }
  if len(a.LastName) > 0 {
    if len(name) > 0 {
      name = name + " "
    }
    name = name + a.LastName
  }
  if len(name) == 0 {
    name = a.UID
  }
  return
}

func (a *Account) Name () (name string) {
  if len(a.FirstName) > 0 {
    name = a.FirstName
  } else if len(a.LastName) > 0 {
    name = a.LastName
  } else {
    name = a.UID
  }
  return
}

func (a *Account) save () (err error) {
  if ! a.dirty {
  } else if _, err = db.InsertUpdate(ACCOUNT_TABLE, a); err == nil {
    a.dirty = false
  }
  return
}

func (a Account) saveNew () (account *Account, err error) {
  a.dirty = true
  if err = (&a).save(); err == nil {
    account = &a
  }
  return
}

func (a *Account) isActive () (bool) {
  return len(a.ActivationCode) == 0
}

func (a *Account) activate (code string) (err error) {
  if a.isActive() {
  } else if code != a.ActivationCode {
    err = ErrInvalidCredentials
  } else {
    a.dirty = true
    a.ActivationCode = ""
  }
  return
}

func New (val func (string) (string)) (account *Account, err error, conflict bool) {
  uid, existing := strings.ToLower(val("uid")), new(Account)
  if e, found := db.Get(ACCOUNT_TABLE, uid, existing); e != nil {
    err = e
  } else if found && existing.isActive() {
    err, conflict = ErrEmailTaken, true
  } else if pwd, e := newPassword(val("pwd1"), val("pwd2")); e != nil {
    err, conflict = e, true
  } else {
    account, err = Account{
      Created: time.Now(),
      UID: uid,
      PWD: pwd,
      Country: val("country"),
      Postcode: strings.ToUpper(val("postcode")),
      FirstName: val("firstname"),
      LastName: val("lastname"),
      ActivationCode: string(util.URLEncode(util.Random(32))),
    }.saveNew()
  }
  return
}

func get (uid string) (account *Account, err error, conflict bool) {
  acc := new(Account)
  if e, found := db.Get(ACCOUNT_TABLE, strings.ToLower(uid), acc); e != nil {
    err = e
  } else if ! found {
    err, conflict = ErrInvalidCredentials, true
  } else {
    account = acc
  }
  return
}

func Activate (uid string, code string) (account *Account, err error, conflict bool) {
  if acc, e, c := get(uid); e != nil {
    err, conflict = e, c
  } else if e := acc.activate(code); e != nil {
    err, conflict = e, true
  } else if e := acc.save(); e != nil {
    err = e
  } else {
    account = acc
  }
  return
}

func Get (uid, pwd string) (account *Account, err error, conflict bool) {
  if acc, e, c := get(uid); e != nil {
    err, conflict = e, c
  } else if ! acc.isActive() {
    err, conflict = ErrNotActivated, true
  } else if e := bcrypt.CompareHashAndPassword(acc.PWD.Value, []byte(pwd)); e != nil {
    err, conflict = ErrInvalidCredentials, true
  } else {
    account = acc
  }
  return
}