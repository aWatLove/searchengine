package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blevesearch/bleve/v2/document"
	"github.com/google/uuid"
	"github.com/valyala/fasthttp"
	"net/http"
	"os"
	"searchengine/internal/common/request"
	"searchengine/internal/config"
	"searchengine/internal/validate"
	"strings"
	"time"
)

var (
	errNotFound         = errors.New("not found")
	errMethodNotAllowed = errors.New("method not allowed")
)

func (s *Server) AddDocumentToIndex(method string, body []byte, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodPost {
		return nil, errMethodNotAllowed
	}

	var doc map[string]interface{}
	err := json.Unmarshal(body, &doc)
	if err != nil {
		return nil, err
	}

	docID := string(args.Peek("docId"))
	if docID == "" {
		docID = uuid.NewString()
	}
	err = s.IndexCli.AddDocument(docID, doc)
	if err != nil {
		return nil, err
	}

	return []byte(docID), nil
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

func (s *Server) getAllDoc(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	resp, err := s.IndexCli.GetAllDoc()
	if err != nil {
		return nil, err
	}

	return json.Marshal(resp)
}

func (s *Server) getDocId(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	docID := string(args.Peek("docId"))

	resp, err := s.IndexCli.GetDocId(docID)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errNotFound
	}

	res := make(map[string]interface{})
	for _, field := range resp.(*document.Document).Fields {
		switch field := field.(type) {
		case *document.TextField:
			res[field.Name()] = string(field.Value())
		case *document.NumericField:
			num, _ := field.Number()
			res[field.Name()] = num
		case *document.DateTimeField:
			dt, _, _ := field.DateTime()
			res[field.Name()] = dt.Format(time.RFC3339)
		case *document.BooleanField:
			b, _ := field.Boolean()
			res[field.Name()] = b
		default:
			res[field.Name()] = field.Value()
		}
	}

	return json.Marshal(res)
}

func (s *Server) GetIndexStruct(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	idxStruct := make(map[string]interface{})
	idxStruct["category"] = []string{""}
	for _, field := range s.Cfg.IndexCfg.Fields {
		switch field.Type {
		case "bool":
			idxStruct[field.Name] = false
		case "number":
			idxStruct[field.Name] = 0
		case "timestamp":
			idxStruct[field.Name] = s.Cfg.DateLayout
		default:
			idxStruct[field.Name] = ""
		}
	}

	return json.MarshalIndent(idxStruct, "", " ")
}

func (s *Server) reindexing(method string, args *fasthttp.Args) error {
	if method != http.MethodGet {
		return errMethodNotAllowed
	}

	return s.IndexCli.ReindexBleve()
}
func (s *Server) rebuildIndex(method string, args *fasthttp.Args) error {
	if method != http.MethodGet {
		return errMethodNotAllowed
	}

	return s.IndexCli.RebuildIndex()
}

func (s *Server) Search(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	query := string(args.Peek("query"))
	if query == "" {
		return nil, errors.New("query is empty")
	}

	var filters *request.FilterRequest
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

	if resp == nil {
		resp = make([]map[string]interface{}, 0)
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

	return json.Marshal(struct {
		Data []string `json:"data"`
	}{category})
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

func (s *Server) getIndexConfig(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	data, err := os.ReadFile(fmt.Sprintf("%s%s", s.Cfg.CfgDirPath, s.Cfg.IndexConfigPath))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Server) getConfigFilter(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	data, err := os.ReadFile(fmt.Sprintf("%s%s", s.Cfg.CfgDirPath, s.Cfg.FilterConfigPath))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Server) getConfigRanking(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	data, err := os.ReadFile(fmt.Sprintf("%s%s", s.Cfg.CfgDirPath, s.Cfg.RankConfigPath))
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Server) updateConfigIndex(method string, body []byte, args *fasthttp.Args) error {
	if method != http.MethodPost {
		return errMethodNotAllowed
	}

	indexCfgNew, err := config.LoadAnyConfigData[*config.IndexConfig](body)
	if err != nil {
		return err
	}

	if s.IndexCli.IsBuilded() {

		tmpIndexPath := fmt.Sprintf("%s%s_old.json", s.Cfg.CfgDirPath, strings.TrimSuffix(s.Cfg.IndexConfigPath, ".json"))
		_ = os.RemoveAll(tmpIndexPath)
		f, err := os.Create(tmpIndexPath)
		if err != nil {
			return err
		}

		dataOldCfg, err := json.MarshalIndent(s.Cfg.IndexCfg, "", " ")
		if err != nil {
			return err
		}

		_, err = f.Write(dataOldCfg)
		if err != nil {
			return err
		}
		f.Close()
	}

	// запись нового конфига
	fNew, err := os.Create(fmt.Sprintf("%s%s", s.Cfg.CfgDirPath, s.Cfg.IndexConfigPath))
	if err != nil {
		return err
	}
	_, err = fNew.Write(body)
	if err != nil {
		return err
	}
	fNew.Close()

	s.IndexCli.SetNeedRebuild()
	s.Cfg.IndexCfg = indexCfgNew

	return nil
}

func (s *Server) isIndexBuilded(method string, args *fasthttp.Args) ([]byte, error) {
	if method != http.MethodGet {
		return nil, errMethodNotAllowed
	}

	return json.Marshal(map[string]interface{}{"isBuilded": s.IndexCli.IsBuilded()})
}

func (s *Server) revertIndexConfig(method string, args *fasthttp.Args) error {
	if method != http.MethodGet {
		return errMethodNotAllowed
	}

	if s.IndexCli.IsBuilded() {
		return errors.New("Can't revert. Index is already builded")
	}

	dataOld, err := os.ReadFile(fmt.Sprintf("%s%s_old.json", s.Cfg.CfgDirPath, strings.TrimSuffix(s.Cfg.IndexConfigPath, ".json")))
	if err != nil {
		return err
	}
	indexCfgOld, err := config.LoadAnyConfigData[*config.IndexConfig](dataOld)
	if err != nil {
		return err
	}

	f, err := os.Create(fmt.Sprintf("%s%s", s.Cfg.CfgDirPath, s.Cfg.IndexConfigPath))
	if err != nil {
		return err
	}
	_, err = f.Write(dataOld)
	if err != nil {
		return err
	}
	defer f.Close()

	_ = os.RemoveAll(fmt.Sprintf("%s%s_old.json", s.Cfg.CfgDirPath, strings.TrimSuffix(s.Cfg.IndexConfigPath, ".json")))

	s.Cfg.IndexCfg = indexCfgOld
	s.IndexCli.SetBuilded()
	return nil
}
