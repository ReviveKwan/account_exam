package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// CloudBaseStore 通过云开发文档库 HTTP API 访问「云开发自带实例」
type CloudBaseStore struct {
	EnvID      string
	APIKey     string
	Instance   string
	Database   string
	HTTPClient *http.Client
	baseURL    string
}

type StoreConfig struct {
	EnvID    string
	APIKey   string
	Instance string
	Database string
}

func ConnectCloudBase(cfg StoreConfig) (*CloudBaseStore, error) {
	if cfg.EnvID == "" {
		return nil, fmt.Errorf("CLOUDBASE_ENV_ID 未配置")
	}
	if cfg.Instance == "" {
		cfg.Instance = "(default)"
	}
	if cfg.Database == "" {
		cfg.Database = "(default)"
	}
	if cfg.APIKey == "" {
		fmt.Println("⚠️  CLOUDBASE_APIKEY 未配置")
	} else {
		fmt.Printf("CloudBase APIKey 已加载，长度=%d\n", len(cfg.APIKey))
	}

	s := &CloudBaseStore{
		EnvID:    cfg.EnvID,
		APIKey:   cfg.APIKey,
		Instance: cfg.Instance,
		Database: cfg.Database,
		HTTPClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		baseURL: fmt.Sprintf(
			"https://%s.api.tcloudbasegateway.com/v1/database/instances/%s/databases/%s",
			cfg.EnvID, cfg.Instance, cfg.Database,
		),
	}
	return s, nil
}

func (s *CloudBaseStore) Ping(ctx context.Context) error {
	_, err := s.Query(ctx, "chapters", map[string]interface{}{}, QueryOpts{Limit: 1})
	return err
}

// EnsureCollections 创建业务集合（已存在则忽略）
func (s *CloudBaseStore) EnsureCollections(ctx context.Context, names ...string) {
	for _, name := range names {
		_, _ = s.doRaw(ctx, http.MethodPost, "/collections", map[string]interface{}{
			"collectionName": name,
		}, nil)
	}
}

type QueryOpts struct {
	Limit  int
	Skip   int
	Sort   map[string]int
	Count  bool
	Fields map[string]int
}

func (s *CloudBaseStore) Query(ctx context.Context, collection string, filter map[string]interface{}, opts QueryOpts) ([]map[string]interface{}, error) {
	if filter == nil {
		filter = map[string]interface{}{}
	}
	q := url.Values{}
	if opts.Limit <= 0 {
		opts.Limit = 1000
	}
	q.Set("limit", strconv.Itoa(opts.Limit))
	q.Set("skip", strconv.Itoa(opts.Skip))
	if opts.Count {
		q.Set("count", "true")
	}
	fb, _ := json.Marshal(filter)
	q.Set("query", string(fb))
	if opts.Sort != nil {
		sb, _ := json.Marshal(opts.Sort)
		q.Set("sort", string(sb))
	}
	if opts.Fields != nil {
		pb, _ := json.Marshal(opts.Fields)
		q.Set("projection", string(pb))
	}

	path := fmt.Sprintf("/collections/%s/documents?%s", collection, q.Encode())
	raw, err := s.doRaw(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	items, err := extractDocList(raw)
	if err != nil {
		return nil, err
	}
	docs := make([]map[string]interface{}, 0, len(items))
	for _, item := range items {
		m, err := decodeDoc(item)
		if err != nil {
			return nil, err
		}
		docs = append(docs, m)
	}
	return docs, nil
}

func (s *CloudBaseStore) Count(ctx context.Context, collection string, filter map[string]interface{}) (int64, error) {
	if filter == nil {
		filter = map[string]interface{}{}
	}
	q := url.Values{}
	q.Set("count", "true")
	q.Set("limit", "1")
	fb, _ := json.Marshal(filter)
	q.Set("query", string(fb))
	path := fmt.Sprintf("/collections/%s/documents?%s", collection, q.Encode())
	raw, err := s.doRaw(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return 0, err
	}
	var resp struct {
		Total int64 `json:"total"`
	}
	if err := json.Unmarshal(raw, &resp); err != nil {
		return 0, err
	}
	return resp.Total, nil
}

func (s *CloudBaseStore) GetByID(ctx context.Context, collection, id string) (map[string]interface{}, error) {
	path := fmt.Sprintf("/collections/%s/documents/%s", collection, url.PathEscape(id))
	raw, err := s.doRaw(ctx, http.MethodGet, path, nil, nil)
	if err != nil {
		return nil, err
	}
	// 单文档接口直接返回文档对象
	var asObj map[string]interface{}
	if err := json.Unmarshal(raw, &asObj); err == nil {
		if _, hasID := asObj["_id"]; hasID {
			return NormalizeEJSON(asObj).(map[string]interface{}), nil
		}
		if data, ok := asObj["data"]; ok {
			b, _ := json.Marshal(data)
			return decodeDoc(b)
		}
	}
	return decodeDoc(raw)
}

func (s *CloudBaseStore) Insert(ctx context.Context, collection string, docs ...map[string]interface{}) error {
	if len(docs) == 0 {
		return nil
	}
	path := fmt.Sprintf("/collections/%s/documents", collection)
	body := map[string]interface{}{"data": docs}
	_, err := s.doRaw(ctx, http.MethodPost, path, body, nil)
	return err
}

func (s *CloudBaseStore) Replace(ctx context.Context, collection, id string, doc map[string]interface{}) error {
	doc["_id"] = id
	body := map[string]interface{}{
		"query":       map[string]interface{}{"_id": id},
		"data":        doc,
		"replaceMode": true,
		"upsert":      true,
	}
	path := fmt.Sprintf("/collections/%s/documents", collection)
	_, err := s.doRaw(ctx, http.MethodPatch, path, body, nil)
	return err
}

func (s *CloudBaseStore) UpdateOne(ctx context.Context, collection, id string, update map[string]interface{}) error {
	path := fmt.Sprintf("/collections/%s/documents/%s", collection, url.PathEscape(id))
	if !hasUpdateOp(update) {
		update = map[string]interface{}{"$set": update}
	}
	// 单文档更新：body 直接是操作符对象，或包一层 data
	body := map[string]interface{}{"data": update}
	_, err := s.doRaw(ctx, http.MethodPatch, path, body, nil)
	return err
}

func (s *CloudBaseStore) UpdateMany(ctx context.Context, collection string, filter, update map[string]interface{}, upsert bool) error {
	if !hasUpdateOp(update) {
		update = map[string]interface{}{"$set": update}
	}
	body := map[string]interface{}{
		"query":  filter,
		"data":   update,
		"upsert": upsert,
	}
	path := fmt.Sprintf("/collections/%s/documents", collection)
	_, err := s.doRaw(ctx, http.MethodPatch, path, body, nil)
	return err
}

func (s *CloudBaseStore) FindOne(ctx context.Context, collection string, filter map[string]interface{}, sort map[string]int) (map[string]interface{}, error) {
	docs, err := s.Query(ctx, collection, filter, QueryOpts{Limit: 1, Sort: sort})
	if err != nil {
		return nil, err
	}
	if len(docs) == 0 {
		return nil, ErrNotFound
	}
	return docs[0], nil
}

func DecodeInto(doc map[string]interface{}, dest interface{}) error {
	raw := NormalizeEJSON(doc)
	normalized, ok := raw.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected document type after ejson normalize")
	}
	if id, ok := normalized["_id"]; ok {
		normalized["id"] = id
	}
	b, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dest)
}

func ToDoc(v interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	if id, ok := m["id"]; ok {
		m["_id"] = id
		delete(m, "id")
	}
	return m, nil
}

var ErrNotFound = fmt.Errorf("document not found")

func (s *CloudBaseStore) do(ctx context.Context, method, path string, body interface{}, out interface{}) error {
	raw, err := s.doRaw(ctx, method, path, body, nil)
	if err != nil {
		return err
	}
	if out == nil || len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, out)
}

func (s *CloudBaseStore) doRaw(ctx context.Context, method, path string, body interface{}, _ interface{}) ([]byte, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, s.baseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	if s.APIKey == "" {
		return nil, fmt.Errorf("CLOUDBASE_APIKEY 为空，请在云托管环境变量配置")
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		// 区分集合不存在 / 文档不存在
		if bytes.Contains(raw, []byte("DOCUMENT_NOT_FOUND")) || bytes.Contains(raw, []byte("not found")) {
			return nil, ErrNotFound
		}
		if bytes.Contains(raw, []byte("DATABASE_COLLECTION_NOT_EXIST")) {
			return nil, fmt.Errorf("集合不存在: %s", truncate(string(raw), 200))
		}
		// 创建集合时 404 较少；查询空集合也可能 200
		if bytes.Contains(raw, []byte("DATABASE_COLLECTION_ALREADY_EXIST")) {
			return raw, nil
		}
		return nil, ErrNotFound
	}
	if resp.StatusCode == http.StatusConflict {
		// 集合已存在
		return raw, nil
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cloudbase http %d: %s", resp.StatusCode, truncate(string(raw), 400))
	}
	return raw, nil
}

func extractDocList(raw []byte) ([]json.RawMessage, error) {
	var wrap struct {
		List []json.RawMessage `json:"list"`
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &wrap); err != nil {
		return nil, err
	}
	if len(wrap.List) > 0 {
		return wrap.List, nil
	}
	return wrap.Data, nil
}

func decodeDoc(raw json.RawMessage) (map[string]interface{}, error) {
	var asObj map[string]interface{}
	if err := json.Unmarshal(raw, &asObj); err == nil {
		return NormalizeEJSON(asObj).(map[string]interface{}), nil
	}
	var asStr string
	if err := json.Unmarshal(raw, &asStr); err == nil {
		if err := json.Unmarshal([]byte(asStr), &asObj); err != nil {
			return nil, err
		}
		return NormalizeEJSON(asObj).(map[string]interface{}), nil
	}
	return nil, fmt.Errorf("cannot decode document: %s", truncate(string(raw), 120))
}

func hasUpdateOp(m map[string]interface{}) bool {
	for k := range m {
		if strings.HasPrefix(k, "$") {
			return true
		}
	}
	return false
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
