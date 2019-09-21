package chiadapter_test

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi"
	chiadapter "github.com/linthan/scf-go-api-proxy/chi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tencentyun/scf-go-lib/events"
)

var _ = Describe("ChiLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")

			r := chi.NewRouter()
			r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("pong"))
			})

			adapter := chiadapter.New(r)

			req := events.APIGatewayRequest{
				Path:   "/ping",
				Method: "GET",
			}

			resp, err := adapter.ProxyWithContext(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))

			resp, err = adapter.Proxy(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})
