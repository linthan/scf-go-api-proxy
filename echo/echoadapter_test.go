package echoadapter_test

import (
	"log"

	"github.com/labstack/echo"
	echoadapter "github.com/linthan/scf-go-api-proxy/echo"
	"github.com/tencentyun/scf-go-lib/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("EchoLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			e := echo.New()
			e.GET("/ping", func(c echo.Context) error {
				log.Println("Handler!!")
				return c.String(200, "pong")
			})

			adapter := echoadapter.New(e)

			req := events.APIGatewayRequest{
				Path:   "/ping",
				Method: "GET",
			}
			resp, err := adapter.Proxy(req)
			Expect(err).To(BeNil())
			Expect(resp.StatusCode).To(Equal(200))
		})
	})
})
