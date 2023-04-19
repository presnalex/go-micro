package rest

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"reflect"
	"strconv"
	"strings"

	"github.com/presnalex/go-micro/v3/wrapper/requestid"
	"go.unistack.org/micro/v3/metadata"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.unistack.org/micro/v3/api"

	"github.com/presnalex/go-micro/v3/logger"
)

var RequestIDHeader = textproto.CanonicalMIMEHeaderKey("X-Request-Id")

var ErrInvalidHandler = errors.New("invalid handler type")

func Register(r *mux.Router, h interface{}, eps []*api.Endpoint) error {
	v := reflect.ValueOf(h)

	methods := v.NumMethod()
	if methods < 1 {
		return ErrInvalidHandler
	}

	for _, ep := range eps {
		idx := strings.Index(ep.Name, ".")
		if idx < 1 || len(ep.Name) <= idx {
			return fmt.Errorf("invalid api.Endpoint name: %s", ep.Name)
		}
		name := ep.Name[idx+1:]
		m := v.MethodByName(name)
		if !m.IsValid() || m.IsZero() {
			return fmt.Errorf("invalid handler, method %s not found", name)
		}

		/*rh, ok := m.Func.Interface().(http.HandlerFunc)
		if !ok {
			return fmt.Errorf("invalid handler, method %s %#+v %#+v not http.Handler", name, m.Func.Interface(), reflect.Indirect(m.Func))
		}
		*/
		rh, ok := m.Interface().(func(http.ResponseWriter, *http.Request))
		if !ok {
			return fmt.Errorf("invalid handler: %#+v", m.Interface())
		}
		r.HandleFunc(ep.Path[0], rh).Methods(ep.Method...).Name(ep.Name)
	}

	r.Use([]mux.MiddlewareFunc{requestIdMiddleware, loggerMiddleware}...)

	return nil
}

func requestIdMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get(RequestIDHeader)
		if id == "" {
			uid, err := uuid.NewRandom()
			if err != nil {
				uid = uuid.Nil
			}
			id = uid.String()
		}
		ctx := logger.InjectLogger(r.Context(), id)
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			md = metadata.New(1)
		}
		_, ok = requestid.GetIncomingRequestId(ctx)
		if !ok {
			md.Set(requestid.DefaultKey, id)
		}
		ctx = metadata.NewIncomingContext(ctx, md)
		rc := r.WithContext(ctx)
		next.ServeHTTP(w, rc)
	})
}

func loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		switch r.URL.Path {
		case "/live", "/ready", "/metrics", "/version":
			next.ServeHTTP(w, r)
			return
		}

		log := logger.FromIncomingContext(r.Context())
		var body []byte
		if r.Body != nil {
			// use http.MaxBytesReader to avoid DoS
			body, _ = ioutil.ReadAll(r.Body)
			// Restore the io.ReadCloser to its original state
			r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		}
		rw := httptest.NewRecorder()
		next.ServeHTTP(rw, r)

		log.Fields(map[string]interface{}{
			"http_method":  r.Method,
			"http_uri":     r.URL.String(),
			"http_reqbody": strconv.Quote(string(body)),
			"http_rspbody": strconv.Quote(string(rw.Body.Bytes())),
			"http_code":    rw.Code,
		}).Info(r.Context(), "")

		// this copies the recorded response to the response writer
		for k, v := range rw.HeaderMap {
			w.Header()[k] = v
		}
		if _, ok := rw.HeaderMap[RequestIDHeader]; !ok {
			if id, ok := requestid.GetIncomingRequestId(r.Context()); ok {
				w.Header()[RequestIDHeader] = []string{id}
			}
		}

		w.WriteHeader(rw.Code)
		rw.Body.WriteTo(w)
	})
}
