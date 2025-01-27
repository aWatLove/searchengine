package server

import (
	"github.com/valyala/fasthttp"
	"searchengine/internal/common/utils"
	"strings"
	"unicode/utf8"
)

const (
	TEST_PATH = "/test"

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
