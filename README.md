## Tencent Function Go Api Proxy

scf-go-api-proxy makes it easy to run Golang APIs written with frameworks such as [Echo](https://echo.labstack.com/) with Tecent Function and Tencent API Gateway.

## Getting started

The first step is to install the required dependencies

```bash
# First, we install the Lambda go libraries
$ go get github.com/tencentyun/scf-go-lib/events
$ go get github.com/tencentyun/scf-go-lib/cloudfunction

# Next, we install the core library
$ go getgithub.com/linthan/scf-go-api-proxy/...
```

Demo

```go
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
			"msg": "ok3",
		})
	})
	echoLambda = echoadapter.New(e)
	cloudfunction.Start(handleRequest)
}
```

## Other frameworks

This package also supports gin and chi

## Deploying the sample

```bash
$ cd scf-lambda-go-api-proxy
$ make
```

The `make` process should generate a `main.zip` file in the sample folder. You can now upload the file to prepare the deployment for Tencent Function and Tencent API Gateway.

```bash
$ cd sample
```

## Deploy

Upload the main.zip to the tencent console

## License

This library is licensed under the Apache 2.0 License.
