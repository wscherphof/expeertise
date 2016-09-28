package secure

import (
	"github.com/julienschmidt/httprouter"
	"github.com/wscherphof/essix/model/account"
	"github.com/wscherphof/essix/ratelimit"
	"github.com/wscherphof/essix/router"
	"github.com/wscherphof/essix/util"
	"github.com/wscherphof/msg"
	"github.com/wscherphof/secure"
	"net/http"
)

func passwordEmail(r *http.Request, acc *account.Account) (error, string) {
	format := msg.Msg(msg.Language(r), "Time format")
	return sendEmail(r, acc.UID, acc.Name(), "password", acc.PasswordCode.Value, acc.PasswordCode.Expires.Format(format))
}

func PasswordCodeForm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err *router.Error) {
	if token, e := ratelimit.NewToken(r); e != nil {
		err = router.NewError(e)
	} else {
		router.Template("secure", "passwordcode", "", map[string]interface{}{
			"UID":            ps.ByName("uid"),
			"RateLimitToken": token,
		})(w, r, ps)
	}
	return
}

func PasswordCode(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err *router.Error) {
	uid := r.FormValue("uid")
	if acc, e, conflict := account.GetInsecure(uid); e != nil {
		err = router.NewError(e)
		err.Conflict = conflict
	} else if !acc.IsActive() {
		err = router.NewError(account.ErrNotActivated, "secure", "activation_resend")
		err.Conflict = true
	} else if e := acc.CreatePasswordCode(); e != nil {
		err = router.NewError(e)
	} else if e, remark := passwordEmail(r, acc); e != nil {
		err = router.NewError(e)
	} else {
		router.Template("secure", "passwordcode_success", "", map[string]interface{}{
			"Name":   acc.Name(),
			"Remark": remark,
		})(w, r, ps)
	}
	return
}

func PasswordForm(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err *router.Error) {
	uid, code, extra, cancel := ps.ByName("uid"), r.FormValue("code"), r.FormValue("extra"), r.FormValue("cancel")
	expires, _ := util.URLDecode([]byte(extra))
	if cancel == "true" {
		account.ClearPasswordCode(uid, code)
		router.Template("secure", "passwordcode_cancelled", "", nil)(w, r, ps)
	} else {
		router.Template("secure", "password", "", map[string]interface{}{
			"UID":     uid,
			"Code":    code,
			"Expires": string(expires),
		})(w, r, ps)
	}
	return
}

func ChangePassword(w http.ResponseWriter, r *http.Request, ps httprouter.Params) (err *router.Error) {
	uid, code, pwd1, pwd2 := r.FormValue("uid"), r.FormValue("code"), r.FormValue("pwd1"), r.FormValue("pwd2")
	if acc, e, conflict := account.GetInsecure(uid); e != nil {
		err = router.NewError(e)
		err.Conflict = conflict
	} else if e, conflict := acc.ChangePassword(code, pwd1, pwd2); e != nil {
		err = router.NewError(e)
		err.Conflict = conflict
	} else {
		secure.LogOut(w, r, false)
		router.Template("secure", "password_success", "", nil)(w, r, ps)
	}
	return
}
