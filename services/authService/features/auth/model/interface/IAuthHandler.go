package _interface

import "github.com/labstack/echo/v4"

type ISigninAuthHandler interface {
	Signin(c echo.Context) error
}

type ISignupAuthHandler interface {
	Signup(c echo.Context) error
}

type IRefreshTokenHandler interface {
	RefreshToken(c echo.Context) error
}

type ILogoutAuthHandler interface {
	Logout(c echo.Context) error
}

type ICheckEmailAuthHandler interface {
	CheckEmail(c echo.Context) error
}

type IGoogleSigninAuthHandler interface{
	GoogleSignin(c echo.Context) error
}