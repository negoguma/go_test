package main

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type User struct {
	Id        string
	AddressId string
}

const VerifyMessage = "verified"

func AuthHandler(next HandlerFunc) HandlerFunc {
	ignore := []string{"/login", "public/index.html"}
	return func(c *Context) {
		//URL prefix가 "/login", "public.index.html"이면 auth를 체크하지 않음
		for _, s := range ignore {
			if strings.HasPrefix(c.Request.URL.Path, s) {
				next(c)
				return
			}
		}
		if v, err := c.Request.Cookie("X_AUTH"); err == http.ErrNoCookie {
			//"X_AUTH" 쿠키 값이 없으면 "/login" 으로 이동
			c.Redirect("/login")
			return
		} else if err != nil {
			//에러 처리
			c.RenderErr(http.StatusInternalServerError, err)
			return
		} else if Verify(VerifyMessage, v.Value) {
			//쿠키 값으로 인증이 확인되면 다음 핸들러로 넘어감
			next(c)
			return
		}

		// "/login"으로 이동
		c.Redirect("/login")
	}
}

func main() {
	// r := &router{make(map[string]map[string]HandlerFunc)}
	//server 생성
	s := NewServer()

	s.HandleFunc("GET", "/login", func(c *Context) {
		//login.html 랜더링
		c.RenderTemplate("/public/login.html", map[string]interface{}{"message": "로그인이 필요합니다"})
	})

	s.HandleFunc("POST", "/login", func(c *Context) {
		//로그인 정보를 확인하여 쿠키에 인증 토큰 값 기록
		if CheckLogin(c.Params["username"].(string), c.Params["password"].(string)) {
			http.SetCookie(c.ResponseWriter, &http.Cookie{
				Name:  "X_AUTH",
				Value: Sign(VerifyMessage),
				Path:  "/",
			})
			c.Redirect("/")
		}
		//id와 password가 맞지 않으면 다시 "/login" 페이지 렌더링
		c.RenderTemplate("/public/login.html", map[string]interface{}{"message": "id 또는 password가 일치하지 않습니다"})
	})

	s.HandleFunc("GET", "/", logHandler(func(c *Context) {
		c.RenderTemplate("/public/index.html", map[string]interface{}{"time": time.Now()})
		fmt.Fprintln(c.ResponseWriter, "welcome!")
	}))

	s.HandleFunc("GET", "/about", logHandler(func(c *Context) {
		fmt.Fprintln(c.ResponseWriter, "about!")
	}))

	s.HandleFunc("GET", "/users/:id", logHandler(recoverHandler(func(c *Context) {
		u := User{Id: c.Params["id"].(string)}
		c.RenderJson(u)
		// if c.Params["id"] == "0" {
		// 	panic("id is zero")
		// }
		// fmt.Fprintln(c.ResponseWriter, "retrieve user %v\n", c.Params["id"])
	})))

	s.HandleFunc("GET", "/users/:user_id/addresses/:address_id", logHandler(func(c *Context) {
		u := User{c.Params["user_id"].(string), c.Params["address_id"].(string)}
		c.RenderJson(u)
		// fmt.Fprintln(c.ResponseWriter, "retrieve user %v's address %v\n", c.Params["user_id"], c.Params["address_id"])
	}))

	s.HandleFunc("POST", "/users", logHandler(recoverHandler(parseFormHandler(parseFormHandler(func(c *Context) {
		fmt.Fprintln(c.ResponseWriter, "create user")
	})))))

	s.HandleFunc("POST", "/users/:user_id/addresses", logHandler(func(c *Context) {
		fmt.Fprintln(c.ResponseWriter, "create user %v's address\n", c.Params["user_id"])
	}))

	s.Use(AuthHandler)
	s.Run(":8080")
}

func CheckLogin(username, password string) bool {
	//로그인 처리
	const (
		USERNAME = "tester"
		PASSWORD = "12345"
	)
	fmt.Print(username, ",", password)
	return username == USERNAME && password == PASSWORD
}

func Verify(message, sig string) bool {
	return hmac.Equal([]byte(sig), []byte(Sign(message)))
}

func Sign(message string) string {
	secretKey := []byte("golang-book-secret-key2")
	if len(secretKey) == 0 {
		return ""
	}
	mac := hmac.New(sha1.New, secretKey)
	io.WriteString(mac, message)
	return hex.EncodeToString(mac.Sum(nil))
}
