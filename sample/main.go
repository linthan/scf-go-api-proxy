package main

import (
	"context"
	"encoding/json"
	"io"

	"github.com/labstack/echo/v4"
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

	e.POST("/hello", func(c echo.Context) error {
		body, _ := io.ReadAll(c.Request().Body)
		return c.JSON(200, map[string]interface{}{
			"msg":  "ok",
			"body": json.RawMessage(body),
		})
	})
	echoLambda = echoadapter.New(e)
	cloudfunction.Start(handleRequest)
}
