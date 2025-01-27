package server

import (
	"errors"
	"github.com/valyala/fasthttp"
	"searchengine/internal/common/utils"
)

var (
	errNotFound         = errors.New("not found")
	errMethodNotAllowed = errors.New("method not allowed")
)

func (s *Server) Handler(path string, ctx *fasthttp.RequestCtx) {

	defer utils.Recovery("SERVER")

	//method, body := string(ctx.Method()), ctx.Request.Body() // todo

	var err error
	var resp []byte
	switch path { //todo
	//case CONST_PATH:
	//	resp, err = s.handlerTest(method, ctx.QueryArgs())
	case TEST_PATH:

	default:
		err = errNotFound
	}

	if err != nil {
		resp = []byte(err.Error())
	}
	setStatusCode(ctx, err)
	if resp != nil {
		ctx.Response.SetBody(resp)
	}
}
