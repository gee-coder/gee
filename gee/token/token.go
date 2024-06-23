package token

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gee-coder/gee"
	"github.com/golang-jwt/jwt/v4"
)

const JWTToken = "gee_token"

type JwtHandler struct {
	// 加密算法
	Alg string
	// 时间函数
	TimeFuc func() time.Time
	// 过期时间
	TimeOut time.Duration
	// 刷新时间
	RefreshTimeOut time.Duration
	// Key
	Key []byte
	// 私钥
	PrivateKey string
	// 刷新key
	RefreshKey string
	// 是否返回给客户端cookie
	SendCookie     bool
	Authenticator  func(ctx *gee.Context) (map[string]any, error)
	CookieName     string
	CookieMaxAge   int64
	CookieDomain   string
	SecureCookie   bool
	CookieHTTPOnly bool
	Header         string
	// 可自定义错误处理Handler
	AuthHandler func(ctx *gee.Context, err error)
}

type JwtResponse struct {
	// 正常用的token
	Token string
	// 过期时间更久的token
	RefreshToken string
}

// 登录  用户认证（用户名密码） -> 用户id 将id生成jwt，并且保存到cookie或者进行返回
func (j *JwtHandler) LoginHandler(ctx *gee.Context) (*JwtResponse, error) {
	data, err := j.Authenticator(ctx)
	if err != nil {
		return nil, err
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	// A部分
	signingMethod := jwt.GetSigningMethod(j.Alg)
	token := jwt.New(signingMethod)
	// B部分
	claims := token.Claims.(jwt.MapClaims)
	if data != nil {
		for key, value := range data {
			claims[key] = value
		}
	}
	if j.TimeFuc == nil {
		j.TimeFuc = func() time.Time {
			return time.Now()
		}
	}
	expire := j.TimeFuc().Add(j.TimeOut)
	// 过期时间
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFuc().Unix()
	var tokenString string
	var tokenErr error
	// C部分 secret
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	jr := &JwtResponse{
		Token: tokenString,
	}
	jr.RefreshToken, err = j.refreshToken(token)
	if err != nil {
		return nil, err
	}
	// 发送cookie
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = expire.Unix() - j.TimeFuc().Unix()
		}
		ctx.SetCookie(j.CookieName, tokenString, int(j.CookieMaxAge), "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}

	return jr, nil
}

func (j *JwtHandler) usingPublicKeyAlgo() bool {
	switch j.Alg {
	case "RS256", "RS512", "RS384":
		return true
	}
	return false
}

func (j *JwtHandler) refreshToken(token *jwt.Token) (string, error) {
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = j.TimeFuc().Add(j.RefreshTimeOut).Unix()
	var tokenString string
	var tokenErr error
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = token.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = token.SignedString(j.Key)
	}
	if tokenErr != nil {
		return "", tokenErr
	}
	return tokenString, nil
}

// LogoutHandler 退出登录
func (j *JwtHandler) LogoutHandler(ctx *gee.Context) error {
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		ctx.SetCookie(j.CookieName, "", -1, "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
		return nil
	}
	return nil
}

// RefreshHandler 刷新token
func (j *JwtHandler) RefreshHandler(ctx *gee.Context) (*JwtResponse, error) {
	rToken, ok := ctx.Get(j.RefreshKey)
	if !ok {
		return nil, errors.New("refresh token is null")
	}
	if j.Alg == "" {
		j.Alg = "HS256"
	}
	// 解析token
	t, err := jwt.Parse(rToken.(string), func(token *jwt.Token) (interface{}, error) {
		if j.usingPublicKeyAlgo() {
			return j.PrivateKey, nil
		} else {
			return j.Key, nil
		}
	})
	if err != nil {
		return nil, err
	}
	// B部分
	claims := t.Claims.(jwt.MapClaims)
	if j.TimeFuc == nil {
		j.TimeFuc = func() time.Time {
			return time.Now()
		}
	}
	expire := j.TimeFuc().Add(j.TimeOut)
	// 过期时间
	claims["exp"] = expire.Unix()
	claims["iat"] = j.TimeFuc().Unix()
	var tokenString string
	var tokenErr error
	// C部分 secret
	if j.usingPublicKeyAlgo() {
		tokenString, tokenErr = t.SignedString(j.PrivateKey)
	} else {
		tokenString, tokenErr = t.SignedString(j.Key)
	}
	if tokenErr != nil {
		return nil, tokenErr
	}
	jr := &JwtResponse{
		Token: tokenString,
	}
	refreshToken, err := j.refreshToken(t)
	if err != nil {
		return nil, err
	}
	jr.RefreshToken = refreshToken
	// 发送存储cookie
	if j.SendCookie {
		if j.CookieName == "" {
			j.CookieName = JWTToken
		}
		if j.CookieMaxAge == 0 {
			j.CookieMaxAge = expire.Unix() - j.TimeFuc().Unix()
		}
		ctx.SetCookie(j.CookieName, tokenString, int(j.CookieMaxAge), "/", j.CookieDomain, j.SecureCookie, j.CookieHTTPOnly)
	}
	return jr, nil
}

// jwt登录中间件
func (j *JwtHandler) AuthInterceptor(next gee.HandlerFunc) gee.HandlerFunc {
	return func(ctx *gee.Context) {
		if j.Header == "" {
			j.Header = "Authorization"
		}
		token := ctx.R.Header.Get(j.Header)
		if token == "" {
			if j.SendCookie {
				cookie, err := ctx.R.Cookie(j.CookieName)
				if err != nil {
					j.AuthErrorHandler(ctx, err)
					return
				}
				token = cookie.String()
			}
		}
		if token == "" {
			j.AuthErrorHandler(ctx, errors.New("token is null"))
			return
		}
		// 解析token
		t, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
			if j.usingPublicKeyAlgo() {
				return j.PrivateKey, nil
			} else {
				return j.Key, nil
			}
		})
		if err != nil {
			j.AuthErrorHandler(ctx, err)
			return
		}
		claims := t.Claims.(jwt.MapClaims)
		fmt.Println(claims)
		ctx.Set("jwt_claims", claims)
		next(ctx)
	}
}

func (j *JwtHandler) AuthErrorHandler(ctx *gee.Context, err error) {
	if j.AuthHandler == nil {
		ctx.W.WriteHeader(http.StatusUnauthorized)
	} else {
		j.AuthHandler(ctx, err)
	}
}
