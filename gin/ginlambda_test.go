package ginadapter_test

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"
	ginadapter "github.com/linthan/scf-go-api-proxy/gin"
	"github.com/tencentyun/scf-go-lib/events"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("GinLambda tests", func() {
	Context("Simple ping request", func() {
		It("Proxies the event correctly", func() {
			log.Println("Starting test")
			r := gin.Default()
			r.GET("/ping", func(c *gin.Context) {
				log.Println("Handler!!")
				c.JSON(200, gin.H{
					"message": "pong",
				})
			})

			adapter := ginadapter.New(r)

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
