package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/forest-shadow/hm-cached-products/internal/application"
	"github.com/forest-shadow/hm-cached-products/internal/logger"
	"github.com/forest-shadow/hm-cached-products/internal/product"
)

type fakeService struct {
	created        product.Input
	statsWindow    string
	updatedVersion int64
}

func newTestLogger(t *testing.T) *logger.Logger {
	t.Helper()

	log, err := logger.New(logger.Config{})
	if err != nil {
		t.Fatalf("create test logger: %v", err)
	}
	return log
}

func TestNewRejectsMissingDependencies(t *testing.T) {
	service := &fakeService{}
	log := newTestLogger(t)

	for _, tc := range []struct {
		name    string
		service productService
		logger  *logger.Logger
		want    string
	}{
		{name: "service", logger: log, want: "httpapi.New: product service is required"},
		{name: "logger", service: service, want: "httpapi.New: logger is required"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			handler, err := New(tc.logger, tc.service)
			if err == nil || err.Error() != tc.want {
				t.Fatalf("error = %v, want %q", err, tc.want)
			}
			if handler != nil {
				t.Fatal("expected nil handler for missing dependency")
			}
		})
	}
}

func TestErrorResponseCodes(t *testing.T) {
	handler, err := New(
		newTestLogger(t),
		&fakeService{},
	)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name        string
		method      string
		path        string
		body        string
		status      int
		wantCode    errorCode
		wantMessage string
	}{
		{name: "invalid JSON", method: http.MethodPost, path: "/products", body: `{`, status: http.StatusBadRequest, wantCode: errorCodeInvalidJSON},
		{name: "malformed trailing data", method: http.MethodPost, path: "/products", body: `{"name":"Tea","price":100}garbage`, status: http.StatusBadRequest, wantCode: errorCodeInvalidJSON},
		{name: "second JSON value", method: http.MethodPost, path: "/products", body: `{"name":"Tea","price":100}{}`, status: http.StatusBadRequest, wantCode: errorCodeMultipleJSON},
		{
			name:     "body too large after first JSON value",
			method:   http.MethodPost,
			path:     "/products",
			body:     `{"name":"Tea","price":100}` + strings.Repeat(" ", maxRequestBodySize),
			status:   http.StatusRequestEntityTooLarge,
			wantCode: errorCodeBodyTooLarge,
		},
		{name: "create missing price", method: http.MethodPost, path: "/products", body: `{"name":"Tea"}`, status: http.StatusBadRequest, wantCode: errorCodeMissingFields, wantMessage: "price is required"},
		{name: "create missing name", method: http.MethodPost, path: "/products", body: `{"price":0}`, status: http.StatusBadRequest, wantCode: errorCodeMissingFields, wantMessage: "name is required"},
		{name: "update missing price", method: http.MethodPut, path: "/products/9", body: `{"name":"Tea","expected_version":1}`, status: http.StatusBadRequest, wantCode: errorCodeMissingFields, wantMessage: "price is required"},
		{name: "update missing name", method: http.MethodPut, path: "/products/9", body: `{"price":0,"expected_version":1}`, status: http.StatusBadRequest, wantCode: errorCodeMissingFields, wantMessage: "name is required"},
		{name: "both fields missing", method: http.MethodPost, path: "/products", body: `{}`, status: http.StatusBadRequest, wantCode: errorCodeMissingFields, wantMessage: "name is required"},
		{name: "invalid ID", method: http.MethodGet, path: "/products/0", status: http.StatusBadRequest, wantCode: errorCodeInvalidID},
		{name: "not found", method: http.MethodGet, path: "/products/10", status: http.StatusNotFound, wantCode: errorCodeProductNotFound},
		{name: "version conflict", method: http.MethodPut, path: "/products/9", body: `{"name":"Tea","price":100,"expected_version":2}`, status: http.StatusConflict, wantCode: errorCodeVersionConflict},
		{name: "invalid stats window", method: http.MethodGet, path: "/products/9/stats?window=bad", status: http.StatusBadRequest, wantCode: errorCodeInvalidStatsWindow},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tc.status {
				t.Fatalf("status = %d, want %d", response.Code, tc.status)
			}
			var body errorResponse
			if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
				t.Fatal(err)
			}
			if body.Code != tc.wantCode {
				t.Errorf("error code = %q, want %q", body.Code, tc.wantCode)
			}
			if body.Message == "" {
				t.Error("error message is empty")
			}
			if tc.wantMessage != "" && body.Message != tc.wantMessage {
				t.Errorf("error message = %q, want %q", body.Message, tc.wantMessage)
			}
		})
	}
}

func (f *fakeService) Create(_ context.Context, input product.Input) (product.Product, error) {
	if input.Price < 0 {
		return product.Product{}, product.ErrInvalid
	}
	f.created = input
	return product.Product{ID: 9, Name: input.Name, Price: input.Price, Version: 1}, nil
}
func (f *fakeService) Get(_ context.Context, id int64) (product.Product, error) {
	if id != 9 {
		return product.Product{}, application.ErrNotFound
	}
	return product.Product{ID: 9, Name: "Example", Price: 100, Version: 1}, nil
}
func (f *fakeService) Update(_ context.Context, command application.UpdateProductCommand) (product.Product, error) {
	id := command.ProductID()
	expectedVersion := command.ExpectedVersion()
	input := command.Input()
	if id != 9 {
		return product.Product{}, application.ErrNotFound
	}
	f.updatedVersion = expectedVersion
	if expectedVersion != 1 {
		return product.Product{}, application.ErrConflict
	}
	return product.Product{ID: id, Name: input.Name, Price: input.Price, Version: 2}, nil
}
func (f *fakeService) Delete(_ context.Context, id int64) error {
	if id != 9 {
		return application.ErrNotFound
	}
	return nil
}
func (f *fakeService) Stats(_ context.Context, id int64, window string) (product.Stats, error) {
	if id != 9 {
		return product.Stats{}, application.ErrNotFound
	}
	if window == "bad" {
		return product.Stats{}, application.ErrInvalidWindow
	}
	f.statsWindow = window
	return product.Stats{ProductID: id, Window: window, ViewCount: 3}, nil
}

func TestEndpoints(t *testing.T) {
	fake := &fakeService{}
	handler, err := New(newTestLogger(t), fake)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, method, path, body string
		wantStatus               int
	}{
		{name: "create product", method: http.MethodPost, path: "/products", body: `{"name":"Example","price":0}`, wantStatus: http.StatusCreated},
		{name: "get product", method: http.MethodGet, path: "/products/9", wantStatus: http.StatusOK},
		{name: "update current version", method: http.MethodPut, path: "/products/9", body: `{"name":"Changed","price":250,"expected_version":1}`, wantStatus: http.StatusOK},
		{name: "reject stale version", method: http.MethodPut, path: "/products/9", body: `{"name":"Changed","price":250,"expected_version":2}`, wantStatus: http.StatusConflict},
		{name: "reject missing version", method: http.MethodPut, path: "/products/9", body: `{"name":"Changed","price":250}`, wantStatus: http.StatusBadRequest},
		{name: "reject zero version", method: http.MethodPut, path: "/products/9", body: `{"name":"Changed","price":250,"expected_version":0}`, wantStatus: http.StatusBadRequest},
		{name: "delete product", method: http.MethodDelete, path: "/products/9", wantStatus: http.StatusNoContent},
		{name: "get seven-day stats", method: http.MethodGet, path: "/products/9/stats?window=7d", wantStatus: http.StatusOK},
		{name: "health endpoint", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK},
		{name: "get missing product", method: http.MethodGet, path: "/products/10", wantStatus: http.StatusNotFound},
		{name: "delete missing product", method: http.MethodDelete, path: "/products/10", wantStatus: http.StatusNotFound},
		{name: "reject zero ID", method: http.MethodGet, path: "/products/0", wantStatus: http.StatusBadRequest},
		{name: "reject unknown stats window", method: http.MethodGet, path: "/products/9/stats?window=bad", wantStatus: http.StatusBadRequest},
		{name: "reject missing price", method: http.MethodPost, path: "/products", body: `{"name":"x"}`, wantStatus: http.StatusBadRequest},
		{name: "reject negative price", method: http.MethodPost, path: "/products", body: `{"name":"x","price":-1}`, wantStatus: http.StatusBadRequest},
		{name: "reject version in create body", method: http.MethodPost, path: "/products", body: `{"name":"x","price":1,"expected_version":1}`, wantStatus: http.StatusBadRequest},
		{name: "reject unknown create field", method: http.MethodPost, path: "/products", body: `{"name":"x","price":1,"unexpected":true}`, wantStatus: http.StatusBadRequest},
		{name: "reject second JSON object", method: http.MethodPost, path: "/products", body: `{"name":"x","price":1}{}`, wantStatus: http.StatusBadRequest},
		{
			name:       "reject oversized body",
			method:     http.MethodPost,
			path:       "/products",
			body:       strings.Repeat(" ", maxRequestBodySize) + `{"name":"x","price":1}`,
			wantStatus: http.StatusRequestEntityTooLarge,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, req)
			if response.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tc.wantStatus, response.Body.String())
			}
			if tc.name == "create product" {
				if response.Header().Get("Location") != "/products/9" {
					t.Fatalf("Location = %q, want /products/9", response.Header().Get("Location"))
				}
				var got product.Product
				if err := json.NewDecoder(bytes.NewReader(response.Body.Bytes())).Decode(&got); err != nil {
					t.Fatalf("decode create response: %v", err)
				}
				if got.Name != "Example" {
					t.Fatalf("created product name = %q, want Example", got.Name)
				}
			}
		})
	}
	if fake.created.Name != "Example" {
		t.Errorf("Create input name = %q, want Example", fake.created.Name)
	}
	if fake.updatedVersion != 2 {
		t.Errorf("last Update expected version = %d, want 2", fake.updatedVersion)
	}
	if fake.statsWindow != "7d" {
		t.Fatalf("stats window = %q, want 7d", fake.statsWindow)
	}
}
