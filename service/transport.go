package service

import (
	"context"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	"net/http"
	"strings"
)

func Index(ctx context.Context, s NetworkAddressService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		response := make(map[string]string)
		response["status"] = "OK"
		encodeResponse(ctx, w, response)
	}
}

func CreateBridgeNetwork(ctx context.Context, svc NetworkAddressService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		netName := getNetworkName(ps)
		createNetworkResponse, err := svc.CreateNetwork(ctx, "bridge", netName)
		encodeResponse(ctx, w, CreateNetworkResponse{Network: createNetworkResponse, Err: err})
	}
}

func CreateOverlayNetwork(ctx context.Context, svc NetworkAddressService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		netName := getNetworkName(ps)
		createNetworkResponse, err := svc.CreateNetwork(ctx, "overlay", netName)
		encodeResponse(ctx, w, CreateNetworkResponse{Network: createNetworkResponse, Err: err})
	}
}

func getNetworkName(ps httprouter.Params) string {
	log.Debug("Request Params: ", ps)
	netName := uuid.NewV4().String()
	if "" != strings.TrimSpace(ps.ByName("name")) {
		netName = strings.TrimSpace(ps.ByName("name"))
	}
	return netName
}

// encodeResponse is the common method to encode all response types to the
// client. I chose to do it this way because, since we're using JSON, there's no
// reason to provide anything more specific. It's certainly possible to
// specialize on a per-response (per-method) basis.
func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(Errorer); ok && e.error() != nil {
		// Provide those as HTTP errors.
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(response)
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	if err == nil {
		panic("encodeError with nil error")
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(codeFrom(err))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}

func codeFrom(err error) int {
	switch err {
	//case ErrNotFound:
	//	return http.StatusNotFound
	//case ErrAlreadyExists, ErrInconsistentIDs:
	//	return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
