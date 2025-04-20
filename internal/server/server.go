package server

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"searchengine/internal/config"
	"searchengine/internal/filter"
	"searchengine/internal/index"
	"searchengine/internal/search"
)

type Server struct {
	HttpServer *fasthttp.Server
	Debug      *http.Server
	Cfg        *config.Config
	IndexCli   *index.Index
	SearchCli  *search.SearchClient
	filterCli  *filter.FilterClient
}

type ServerPrivate struct {
	HttpServer *http.Server
}

func New(cfg *config.Config, indxCli *index.Index, searchCli *search.SearchClient, filterCli *filter.FilterClient) *Server {
	return &Server{
		HttpServer: new(fasthttp.Server),
		Debug: &http.Server{
			Addr: cfg.PrivatePort,
		},
		Cfg:       cfg,
		IndexCli:  indxCli,
		SearchCli: searchCli,
		filterCli: filterCli,
	}
}

func NewServerPrivate(port string) *ServerPrivate {
	return &ServerPrivate{
		HttpServer: &http.Server{
			Addr: port,
		},
	}
}

func (s *Server) Start() {
	s.HttpServer.Handler = corsMiddleware(jsonMiddleware(s.Router))
	s.Debug.Handler = s.initRoutsServerPrivate()

	go func() {
		if err := s.HttpServer.ListenAndServe(
			fmt.Sprintf("0.0.0.0%s", s.Cfg.PublicPort),
		); err != nil {
			log.Fatal(err)
		}
	}()

	go func() {
		if err := s.Debug.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()
}

func setStatusCode(ctx *fasthttp.RequestCtx, err error) {
	if err != nil {
		switch err {
		case errNotFound:
			ctx.SetStatusCode(fasthttp.StatusNotFound)
		case errMethodNotAllowed:
			ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
		default:
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		}
	} else {
		ctx.SetStatusCode(fasthttp.StatusOK)
	}
}

func (s *Server) initRoutsServerPrivate() http.Handler {
	privateMux := http.NewServeMux()
	privateMux.Handle("/metrics", promhttp.Handler())
	privateMux.HandleFunc("/health", func(writer http.ResponseWriter, request *http.Request) { writer.WriteHeader(http.StatusOK) })

	return privateMux
}

func (s *Server) Stop(ctx context.Context) error {
	err := s.Debug.Shutdown(ctx)
	if err == http.ErrServerClosed {
		log.Println("[SERVER][STOP] HTTP private server stopped")
	} else {
		return err
	}

	err = s.HttpServer.Shutdown()
	if err == http.ErrServerClosed {
		log.Println("[SERVER][STOP] HTTP public server stopped")
		return nil
	} else {
		return err
	}
}

func corsMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Разрешаем запросы с любого origin (для продакшена укажите конкретные домены)
		ctx.Response.Header.Set("Access-Control-Allow-Origin", "*")

		// Разрешаем методы
		ctx.Response.Header.Set("Access-Control-Allow-Methods",
			"GET, POST, HEAD, PUT, DELETE, PATCH, OPTIONS")

		// Разрешаем заголовки
		ctx.Response.Header.Set("Access-Control-Allow-Headers",
			"Origin, Content-Type, Accept, Authorization, X-Requested-With")

		// Разрешаем куки и авторизацию (если нужно)
		// ctx.Response.Header.Set("Access-Control-Allow-Credentials", "true")

		// Максимальное время кеширования предварительных запросов
		ctx.Response.Header.Set("Access-Control-Max-Age", "3600")

		// Обработка предварительного OPTIONS-запроса
		if string(ctx.Method()) == "OPTIONS" {
			ctx.SetStatusCode(fasthttp.StatusNoContent)
			return
		}

		// Передаем запрос следующему обработчику
		next(ctx)
	}
}

func jsonMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Set("Content-Type", "application/json")
		next(ctx)
	}
}
