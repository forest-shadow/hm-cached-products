package httpapi

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/forest-shadow/hm-cached-products/internal/application"
	"github.com/forest-shadow/hm-cached-products/internal/logger"
	"github.com/forest-shadow/hm-cached-products/internal/product"
)

type productService interface {
	Create(ctx context.Context, input product.Input) (product.Product, error)
	Get(ctx context.Context, id int64) (product.Product, error)
	Update(ctx context.Context, command application.UpdateProductCommand) (product.Product, error)
	Delete(ctx context.Context, id int64) error
	Stats(ctx context.Context, id int64, window string) (product.Stats, error)
}

type Handler struct {
	logger  *logger.Logger
	service productService
}

func New(logger *logger.Logger, service productService) (http.Handler, error) {
	const op = "httpapi.New"

	if logger == nil {
		return nil, fmt.Errorf("%s: logger is required", op)
	}
	if service == nil {
		return nil, fmt.Errorf("%s: product service is required", op)
	}
	handler := &Handler{logger: logger, service: service}
	return handler.routes(), nil
}

func (h *Handler) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /products", h.create)
	mux.HandleFunc("GET /products/{id}", h.get)
	mux.HandleFunc("PUT /products/{id}", h.update)
	mux.HandleFunc("DELETE /products/{id}", h.delete)
	mux.HandleFunc("GET /products/{id}/stats", h.stats)

	mux.HandleFunc("GET /health", h.health)
	return mux
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, errInvalidID
	}
	return id, nil
}

func (h *Handler) create(w http.ResponseWriter, r *http.Request) {
	request, err := decodeJSON[createProductRequest](w, r)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	input, err := request.toProductInput()
	if err != nil {
		writeRequestError(w, err)
		return
	}
	p, err := h.service.Create(r.Context(), input)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.Header().Set("Location", "/products/"+strconv.FormatInt(p.ID, 10))
	writeJSON(w, http.StatusCreated, productResponse(p))
}

func (h *Handler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeRequestError(w, err)
		return
	}
	p, err := h.service.Get(r.Context(), id)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, productResponse(p))
}

func (h *Handler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeRequestError(w, err)
		return
	}
	request, err := decodeJSON[updateProductRequest](w, r)
	if err != nil {
		writeDecodeError(w, err)
		return
	}
	params, err := request.toUpdateProductParams()
	if err != nil {
		writeRequestError(w, err)
		return
	}
	command, err := application.NewUpdateProductCommand(
		id,
		params.ExpectedVersion,
		params.Input,
	)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	p, err := h.service.Update(r.Context(), command)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, productResponse(p))
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeRequestError(w, err)
		return
	}
	if err := h.service.Delete(r.Context(), id); err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r)
	if err != nil {
		writeRequestError(w, err)
		return
	}
	window := r.URL.Query().Get("window")
	stats, err := h.service.Stats(r.Context(), id, window)
	if err != nil {
		h.writeServiceError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, statsResponse(stats))
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}
