package server

import (
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"net/http"
	"searchengine/internal/common/request"
	"searchengine/internal/validate"
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
	if method != http.MethodDelete {
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

	var filters *request.FilterRequest //todo
	filtersData := args.Peek("filters")
	if len(filtersData) != 0 {
		err := json.Unmarshal(filtersData, &filters)
		if err != nil {
			return nil, err
		}
	}

	sortField := string(args.Peek("sortField"))
	if sortField != "" {
		if !validate.ValidateSortField(s.Cfg, sortField) {
			return nil, errors.New("invalid sort field")
		}
	}
	sortOrder := string(args.Peek("sortOrder"))

	resp, err := s.SearchCli.AdvancedSearch(query, filters, sortField, sortOrder)
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

func (s *Server) FiltersByCategory(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	category := string(args.Peek("category"))
	if category == "" {
		return nil, errors.New("category is empty")
	}

	filters, ok := s.filterCli.GetByCategory(category)
	if !ok {
		return nil, errNotFound
	}
	return json.Marshal(&filters)
}

func (s *Server) GetAllCategories(method string) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	category := s.filterCli.GetAllCategories()

	return json.Marshal(&category)
}
