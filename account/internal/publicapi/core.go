package publicapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"pulse-auth/internal/service/user"
	"pulse-auth/internal/utils"

	"github.com/go-http-utils/headers"
	"go.uber.org/zap"
)

type Handler struct {
	Logger      *zap.Logger
	UserService user.Service
}

func (h *Handler) writeError(ctx context.Context, w http.ResponseWriter, err error) {
	h.Logger.Sugar().Warnf("http response error: %v", err)

	w.Header().Set(headers.ContentType, "application/json")
	errorResult, ok := utils.FromError(err)
	if !ok {
		h.Logger.Error("cannot write log message")
		return
	}
	w.WriteHeader(errorResult.StatusCode)
	err = json.NewEncoder(w).Encode(
		map[string]any{
			"message": errorResult.Msg,
			"code":    errorResult.StatusCode,
		})

	if err != nil {
		http.Error(w, utils.InternalErrorMessage, http.StatusInternalServerError)
	}
}

func writeResponse(w http.ResponseWriter, response any) {
	w.Header().Set(headers.ContentType, "application/json")
	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		http.Error(w, utils.InternalErrorMessage, http.StatusInternalServerError)
	}
}

func parseJSONRequest[T loginRequest | registerRequest](r *http.Request) (*T, error) {
	defer r.Body.Close()
	body, err := io.ReadAll(r.Body)
	if err != nil {
		err = fmt.Errorf("read body: %w", err)
		return nil, utils.WrapInternalError(err)
	}

	var request T
	err = json.Unmarshal(body, &request)
	if err != nil {
		err = fmt.Errorf("unmarshal request body: %w", err)
		return nil, utils.WrapValidationError(err)
	}
	return &request, nil
}
