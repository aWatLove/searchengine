package server

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"net/http"
)

var (
	errNotFound         = errors.New("not found")
	errMethodNotAllowed = errors.New("method not allowed")
)

func (s *Server) AddDocumentToIndex(method string, body []byte, args *fasthttp.Args) error {
	if method != http.MethodPost {
		return errMethodNotAllowed
	}

	var doc map[string]interface{}
	err := json.Unmarshal(body, &doc)
	if err != nil {
		return err
	}

	docID := string(args.Peek("docId"))
	if docID == "" {
		docID = uuid.NewString()
	}
	err = s.IndexCli.AddDocument(docID, doc)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) UpdateDocument(method string, body []byte, args *fasthttp.Args) error {
	if method != http.MethodPost {
		return errMethodNotAllowed
	}

	var doc map[string]interface{}
	err := json.Unmarshal(body, &doc)
	if err != nil {
		return err
	}

	docID := string(args.Peek("docId"))
	if docID == "" {
		return errors.New("docId is empty")
	}

	err = s.IndexCli.Update(docID, doc)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) DeleteDocument(method string, args *fasthttp.Args) error {
	if method != http.MethodPost {
		return errMethodNotAllowed
	}

	docID := string(args.Peek("docId"))
	if docID == "" {
		return errors.New("docId is empty")
	}

	err := s.IndexCli.Delete(docID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) Search(method string, args *fasthttp.Args) ([]byte, error) { //todo
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	query := string(args.Peek("query"))
	if query == "" {
		return nil, errors.New("query is empty")
	}

	filters := make(map[string]interface{}, 0) //todo
	sorts := make([]string, 0)                 //todo

	resp, err := s.SearchCli.SearchIndex(query, filters, sorts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&resp)
}

func (s *Server) SimpleSearch(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	query := string(args.Peek("query"))
	if query == "" {
		return nil, errors.New("query is empty")
	}

	filters := make(map[string]interface{}, 0) //todo
	sorts := make([]string, 0)                 //todo

	resp, err := s.SearchCli.SearchIndex(query, filters, sorts)
	if err != nil {
		return nil, err
	}

	return json.Marshal(&resp)
}
