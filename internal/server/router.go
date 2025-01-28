package server

import (
	"github.com/valyala/fasthttp"
	"searchengine/internal/common/utils"
	"strings"
	"unicode/utf8"
)

const (
	TEST_PATH = "/test"

	// INDEX
	ADD_DOCUMENT_TO_INDEX_PATH      = "/addDoc"
	UPDATE_DOCUMENT_IN_INDEX_PATH   = "/updateDoc"
	DELETE_DOCUMENT_FROM_INDEX_PATH = "/deleteDoc"

	// SEARCH
	SEARCH_SIMPLE_PATH = "/simpleSearch"
	SEARCH_PATH        = "/search"

	V1 = "/api/v1"
)

func (s *Server) Router(ctx *fasthttp.RequestCtx) {
	defer utils.Recovery("ROUTER")

	path := string(ctx.Path())
	if !utf8.ValidString(path) {
		return
	}

	switch {
	case strings.Contains(path, V1):
		//metrics.IncRpsHandle(path)
		s.Handler(strings.TrimPrefix(path, V1), ctx)
	}

}

func (s *Server) Handler(path string, ctx *fasthttp.RequestCtx) {

	defer utils.Recovery("SERVER")

	method, body := string(ctx.Method()), ctx.Request.Body()

	var err error
	var resp []byte
	switch path {
	// INDEX
	case ADD_DOCUMENT_TO_INDEX_PATH:
		err = s.AddDocumentToIndex(method, body, ctx.QueryArgs())
	case UPDATE_DOCUMENT_IN_INDEX_PATH:
		err = s.UpdateDocument(method, body, ctx.QueryArgs())
	case DELETE_DOCUMENT_FROM_INDEX_PATH:
		err = s.DeleteDocument(method, ctx.QueryArgs())

	// SEARCH
	case SEARCH_PATH:
		resp, err = s.Search(method, ctx.QueryArgs())
	case SEARCH_SIMPLE_PATH:
		resp, err = s.SimpleSearch(method, ctx.QueryArgs())

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
