package secure

import (
	"github.com/wscherphof/expeertise/ratelimit"
	"github.com/wscherphof/expeertise/router"
)

func init() {
	router.GET("/account", IfSecureHandle(UpdateAccountForm, SignUpForm))
	router.POST("/account", ratelimit.Handle(3600, SignUp))
	router.PUT("/account", SecureHandle(UpdateAccount))

	router.GET("/session", LogInForm)
	router.POST("/session", ratelimit.Handle(60, LogIn))
	router.DELETE("/session", LogOut)

	router.GET("/account/activation/:uid", ActivateForm)
	router.GET("/account/activation", ActivateForm)
	router.PUT("/account/activation", Activate)
	router.GET("/account/activationcode/:uid", ActivationCodeForm)
	router.GET("/account/activationcode", ActivationCodeForm)
	router.POST("/account/activationcode", ActivationCode)

	router.GET("/account/passwordcode/:uid", PasswordCodeForm)
	router.GET("/account/passwordcode", PasswordCodeForm)
	router.POST("/account/passwordcode", PasswordCode)
	router.GET("/account/password/:uid", PasswordForm)
	router.PUT("/account/password", ChangePassword)

	router.GET("/account/emailaddresscode", SecureHandle(EmailAddressCodeForm))
	router.POST("/account/emailaddresscode", SecureHandle(EmailAddressCode))
	router.GET("/account/emailaddress/*filepath", SecureHandle(EmailAddressForm))
	router.PUT("/account/emailaddress", SecureHandle(ChangeEmailAddress))

	router.GET("/account/terminatecode", SecureHandle(TerminateCodeForm))
	router.POST("/account/terminatecode", SecureHandle(TerminateCode))
	router.GET("/account/terminate/*filepath", SecureHandle(TerminateForm))
	router.DELETE("/account", SecureHandle(Terminate))
}
