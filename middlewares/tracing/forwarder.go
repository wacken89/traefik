package tracing

import (
	"fmt"
	"net/http"

	"github.com/containous/traefik/log"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/urfave/negroni"
)

type forwarderMiddleware struct {
	frontend string
	backend  string
	opName   string
	*Tracing
}

// NewForwarderMiddleware creates a new forwarder middleware that traces the outgoing request
func (t *Tracing) NewForwarderMiddleware(frontend, backend string) negroni.Handler {
	log.Debugf("Added outgoing tracing middleware %s", frontend)
	return &forwarderMiddleware{
		Tracing:  t,
		frontend: frontend,
		backend:  backend,
		opName:   fmt.Sprintf("forward %s/%s", frontend, backend),
	}
}

func (f *forwarderMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	span, r, finish := StartSpan(r, f.opName, true)
	defer finish()
	span.SetTag("frontend.name", f.frontend)
	span.SetTag("backend.name", f.backend)
	ext.HTTPMethod.Set(span, r.Method)
	ext.HTTPUrl.Set(span, r.URL.String())
	span.SetTag("http.host", r.Host)

	InjectRequestHeaders(r)

	w = &statusCodeTracker{w, 200}

	next(w, r)

	LogResponseCode(span, w.(*statusCodeTracker).status)
}
