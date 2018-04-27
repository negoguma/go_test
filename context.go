package main

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"net/http"
	"path/filepath"
)

type Context struct {
	Params map[string]interface{}

	ResponseWriter http.ResponseWriter
	Request        *http.Request
}

type HandlerFunc func(*Context)

//templates: 템플릿 객체를 보관하기 위한 map
var templates = map[string]*template.Template{}

func (c *Context) Redirect(url string) {
	http.Redirect(c.ResponseWriter, c.Request, url, http.StatusMovedPermanently)
}

func (c *Context) RenderTemplate(path string, v interface{}) {
	//path에 해당하는 템플릿이 있는지 확인
	t, ok := templates[path]
	if !ok {
		//path에 해당하는 템프릿이 없으면 템플릿이 없으면 템플릿 객체 생성
		t = template.Must(template.ParseFiles(filepath.Join(".", path)))
		templates[path] = t
	}

	//v값을 템플릿 내부로 전달하여 만들어진 최종 결과를 c.ResponseWriter에 출력
	t.Execute(c.ResponseWriter, v)
}

func (c *Context) RenderJson(v interface{}) {
	//HTTP status를 StatusOK 로 지정
	c.ResponseWriter.WriteHeader(http.StatusOK)
	//Content-Type을 apliication/json으로 지정
	c.ResponseWriter.Header().Set("Content-Type", "application/json; charset=utf-8")

	//v 값을 json으로 출력
	if err := json.NewEncoder(c.ResponseWriter).Encode(v); err != nil {
		//에러 발생 시 RenderErr메서드 호출
		c.RenderErr(http.StatusInternalServerError, err)
	}
}

func (c *Context) RenderXml(v interface{}) {
	//HTTP status를 StatusOK 로 지정
	c.ResponseWriter.WriteHeader(http.StatusOK)
	//Content-Type을 apliication/xml 지정
	c.ResponseWriter.Header().Set("Content-Type", "application/xml; charset=utf-8")

	//v 값을 xml 출력
	if err := xml.NewEncoder(c.ResponseWriter).Encode(v); err != nil {
		//에러 발생 시 RenderErr메서드 호출
		c.RenderErr(http.StatusInternalServerError, err)
	}
}

func (c *Context) RenderErr(code int, err error) {
	if err != nil {
		if code > 0 {
			//정상적인 code를 전달하면 HTTP Status를 해당 code로 지정
			http.Error(c.ResponseWriter, http.StatusText(code), code)
		} else {
			//정상적인 code가 아니면HTTP Status를 StatusInternalServerError로 지정
			defaultErr := http.StatusInternalServerError
			http.Error(c.ResponseWriter, http.StatusText(defaultErr), defaultErr)
		}
	}
}
