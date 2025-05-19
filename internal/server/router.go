package server

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"searchengine/internal/common/utils"
	"searchengine/internal/metrics"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	TEST_PATH = "/test"

	// INDEX
	ADD_DOCUMENT_TO_INDEX_PATH      = "/addDoc"
	UPDATE_DOCUMENT_IN_INDEX_PATH   = "/updateDoc"
	DELETE_DOCUMENT_FROM_INDEX_PATH = "/deleteDoc"
	REINDEX_PATH                    = "/reindex"
	GET_INDEX_STRUCT                = "/indexStruct"
	REBUILD_INDEX_PATH              = "/rebuild"

	GET_ALL_DOCUMENTS = "/getAllDoc"
	GET_DOCUMENT_ID   = "/getDocId"

	// SEARCH
	SEARCH_SIMPLE_PATH = "/simpleSearch"
	SEARCH_PATH        = "/search"

	// FILTERS
	FILTERS_BY_CATEGORY      = "/filtersByCategory"
	FILTERS_GET_ALL_CATEGORY = "/category"

	// CONFIGS
	GET_CONFIG_INDEX_PATH   = "/getConfig/index"
	GET_CONFIG_FILTER_PATH  = "/getConfig/filter"
	GET_CONFIG_RANKING_PATH = "/getConfig/ranking"

	UPD_CONFIG_INDEX_PATH    = "/config/index"
	REVERT_CONFIG_INDEX_PATH = "/config/index/revert"
	INDEX_IS_BUILD_PATH      = "/config/index/isbuild"

	UPD_CONFIG_FILTER_PATH  = "/config/filter"
	UPD_CONFIG_RANKING_PATH = "/config/ranking"

	// LOGS
	LAST_LOG_PATH  = "/lastlog"
	LIST_LOGS_PATH = "/listlogs"
	LOG_PATH       = "/log"

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
	t := time.Now()
	defer utils.Recovery("SERVER")

	method, body := string(ctx.Method()), ctx.Request.Body()

	var err error
	var resp []byte
	switch path {
	// INDEX
	case ADD_DOCUMENT_TO_INDEX_PATH:
		resp, err = s.AddDocumentToIndex(method, body, ctx.QueryArgs())
	case UPDATE_DOCUMENT_IN_INDEX_PATH:
		err = s.UpdateDocument(method, body, ctx.QueryArgs())
	case DELETE_DOCUMENT_FROM_INDEX_PATH:
		err = s.DeleteDocument(method, ctx.QueryArgs())
	case GET_ALL_DOCUMENTS:
		resp, err = s.getAllDoc(method, ctx.QueryArgs())
	case REINDEX_PATH:
		err = s.reindexing(method, ctx.QueryArgs())
	case GET_INDEX_STRUCT:
		resp, err = s.GetIndexStruct(method, ctx.QueryArgs())
	case GET_DOCUMENT_ID:
		resp, err = s.getDocId(method, ctx.QueryArgs())
	case REBUILD_INDEX_PATH:
		err = s.rebuildIndex(method, ctx.QueryArgs())

	// SEARCH
	case SEARCH_PATH:
		resp, err = s.Search(method, ctx.QueryArgs())
	case SEARCH_SIMPLE_PATH:
		resp, err = s.SimpleSearch(method, ctx.QueryArgs())

	// FILTERS
	case FILTERS_BY_CATEGORY:
		resp, err = s.FiltersByCategory(method, ctx.QueryArgs())
	case FILTERS_GET_ALL_CATEGORY:
		resp, err = s.GetAllCategories(method)

	// CONFIGS
	case GET_CONFIG_INDEX_PATH:
		resp, err = s.getIndexConfig(method, ctx.QueryArgs())
	case UPD_CONFIG_INDEX_PATH:
		err = s.updateConfigIndex(method, body, ctx.QueryArgs())
	case REVERT_CONFIG_INDEX_PATH:
		err = s.revertIndexConfig(method, ctx.QueryArgs())
	case INDEX_IS_BUILD_PATH:
		resp, err = s.isIndexBuilded(method, ctx.QueryArgs())

	case GET_CONFIG_FILTER_PATH:
		resp, err = s.getConfigFilter(method, ctx.QueryArgs())
	case UPD_CONFIG_FILTER_PATH:
		err = s.updateConfigFilter(method, body, ctx.QueryArgs())

	case GET_CONFIG_RANKING_PATH:
		resp, err = s.getConfigRanking(method, ctx.QueryArgs())
	case UPD_CONFIG_RANKING_PATH:
		err = s.updateConfigRanking(method, body, ctx.QueryArgs())
	case LAST_LOG_PATH:
		_, err = s.lastLogHandler(ctx)
	case LIST_LOGS_PATH:
		resp, err = s.listLogsHandler(ctx)
	case LOG_PATH:
		s.logHandler(ctx)
	case "/metrics":
		promhttp.Handler()

	default:
		err = errNotFound
	}

	if err != nil {
		metrics.ErrorRPS(path, method)
		resp = []byte(err.Error())
	}
	setStatusCode(ctx, err)
	if resp != nil {
		ctx.Response.SetBody(resp)
	}

	metrics.RequestRPS(path, method, fmt.Sprintf("%d", ctx.Response.StatusCode()))
	metrics.RequestDuration(path, method, time.Since(t).Seconds())
}
