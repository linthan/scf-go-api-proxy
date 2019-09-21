package main

import (
	"context"

	"github.com/labstack/echo"
	echoadapter "github.com/linthan/scf-go-api-proxy/echo"
	"github.com/tencentyun/scf-go-lib/cloudfunction"
	"github.com/tencentyun/scf-go-lib/events"
)

var echoLambda *echoadapter.EchoLambda

func handleRequest(ctx context.Context, request events.APIGatewayRequest) (events.APIGatewayResponse, error) {
	return echoLambda.ProxyWithContext(ctx, request)
}

func main() {
	e := echo.New()
	e.GET("/hello", func(c echo.Context) error {
		return c.JSON(200, map[string]interface{}{
			"msg": "ok",
		})
	})
	echoLambda = echoadapter.New(e)
	cloudfunction.Start(handleRequest)
}
