package main

import "net/http"

type Server struct {
	*router
	middleware   []Middleware
	startHandler HandlerFunc
}

func NewServer() *Server {
	r := &router{make(map[string]map[string]HandlerFunc)}
	s := &Server{router: r}
	s.middleware = []Middleware{
		logHandler,
		recoverHandler,
		staticHandler,
		parseFormHandler,
		paresJsonBodyHandler}
	return s
}

func (s *Server) Use(middleware ...Middleware) {
	s.middleware = append(s.middleware, middleware...)
}

func (s *Server) Run(addr string) {
	//startHandler를 라우터 핸들러 함수로 지정
	s.startHandler = s.router.handler()

	//등록된 미들웨어를 라우터 핸들러 앞에 하나씩 추가
	for i := len(s.middleware) - 1; i >= 0; i-- {
		s.startHandler = s.middleware[i](s.startHandler)
	}

	//웹서버 시작
	if err := http.ListenAndServe(addr, s); err != nil {
		panic(err)
	}
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//Context 생성
	c := &Context{
		Params:         make(map[string]interface{}),
		ResponseWriter: w,
		Request:        r,
	}
	for k, v := range r.URL.Query() {
		c.Params[k] = v[0]
	}
	s.startHandler(c)
}
