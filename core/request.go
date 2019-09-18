// Package core provides utility methods that help convert proxy events
// into an http.Request and http.ResponseWriter
package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/tencentyun/scf-go-lib/events"
	"github.com/tencentyun/scf-go-lib/functioncontext"
)

// CustomHostVariable is the name of the environment variable that contains
// the custom hostname for the request. If this variable is not set the framework
// reverts to `DefaultServerAddress`. The value for a custom host should include
// a protocol: http://my-custom.host.com
const CustomHostVariable = "GO_API_HOST"

// DefaultServerAddress is prepended to the path of each incoming reuqest
const DefaultServerAddress = "https://tencent-serverless-go-api.com"

// APIGwContextHeader is the custom header key used to store the
// API Gateway context. To access the Context properties use the
// GetAPIGatewayContext method of the RequestAccessor object.
const APIGwContextHeader = "X-GoLambdaProxy-ApiGw-Context"

// RequestAccessor objects give access to custom API Gateway properties
// in the request.
type RequestAccessor struct {
	stripBasePath string
}

// GetAPIGatewayContext extracts the API Gateway context object from a
// request's custom header.
// Returns a populated events.APIGatewayProxyRequestContext object from
// the request.
func (r *RequestAccessor) GetAPIGatewayContext(req *http.Request) (events.APIGatewayRequestContext, error) {
	if req.Header.Get(APIGwContextHeader) == "" {
		return events.APIGatewayRequestContext{}, errors.New("No context header in request")
	}
	context := events.APIGatewayRequestContext{}
	err := json.Unmarshal([]byte(req.Header.Get(APIGwContextHeader)), &context)
	if err != nil {
		log.Println("Erorr while unmarshalling context")
		log.Println(err)
		return events.APIGatewayRequestContext{}, err
	}
	return context, nil
}

// StripBasePath instructs the RequestAccessor object that the given base
// path should be removed from the request path before sending it to the
// framework for routing. This is used when API Gateway is configured with
// base path mappings in custom domain names.
func (r *RequestAccessor) StripBasePath(basePath string) string {
	if strings.Trim(basePath, " ") == "" {
		r.stripBasePath = ""
		return ""
	}

	newBasePath := basePath
	if !strings.HasPrefix(newBasePath, "/") {
		newBasePath = "/" + newBasePath
	}

	if strings.HasSuffix(newBasePath, "/") {
		newBasePath = newBasePath[:len(newBasePath)-1]
	}

	r.stripBasePath = newBasePath

	return newBasePath
}

// ProxyEventToHTTPRequest converts an API Gateway proxy event into a http.Request object.
// Returns the populated http request with additional two custom headers for the stage variables and API Gateway context.
// To access these properties use the GetAPIGatewayStageVars and GetAPIGatewayContext method of the RequestAccessor object.
func (r *RequestAccessor) ProxyEventToHTTPRequest(req events.APIGatewayRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToHeader(httpRequest, req)
}

// EventToRequestWithContext converts an API Gateway proxy event and context into an http.Request object.
// Returns the populated http request with lambda context, stage variables and APIGatewayProxyRequestContext as part of its context.
// Access those using GetAPIGatewayContextFromContext, GetStageVarsFromContext and GetRuntimeContextFromContext functions in this package.
func (r *RequestAccessor) EventToRequestWithContext(ctx context.Context, req events.APIGatewayRequest) (*http.Request, error) {
	httpRequest, err := r.EventToRequest(req)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return addToContext(ctx, httpRequest, req), nil
}

// EventToRequest converts an API Gateway proxy event into an http.Request object.
// Returns the populated request maintaining headers
func (r *RequestAccessor) EventToRequest(req events.APIGatewayRequest) (*http.Request, error) {
	decodedBody := ioutil.NopCloser(bytes.NewBufferString(req.Body))
	path := req.Path
	if r.stripBasePath != "" && len(r.stripBasePath) > 1 {
		if strings.HasPrefix(path, r.stripBasePath) {
			path = strings.Replace(path, r.stripBasePath, "", 1)
		}
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	serverAddress := DefaultServerAddress
	if customAddress, ok := os.LookupEnv(CustomHostVariable); ok {
		serverAddress = customAddress
	}
	path = serverAddress + path

	if len(req.QueryString) > 0 {
		queryString := ""
		for q := range req.QueryString {
			if queryString != "" {
				queryString += "&"
			}
			queryString += url.QueryEscape(q) + "=" + url.QueryEscape(req.QueryString[q][0])
		}
		path += "?" + queryString
	}

	httpRequest, err := http.NewRequest(
		strings.ToUpper(req.Method),
		path,
		decodedBody,
	)

	if err != nil {
		fmt.Printf("Could not convert request %s:%s to http.Request\n", req.Method, req.Path)
		log.Println(err)
		return nil, err
	}
	for h := range req.Headers {
		httpRequest.Header.Add(h, req.Headers[h])
	}
	httpRequest.Header.Add(http.CanonicalHeaderKey("x-apigateway-serviceid"), req.Context.ServiceID)
	httpRequest.Header.Add(http.CanonicalHeaderKey("x-apigateway-requestid"), req.Context.RequestID)
	httpRequest.Header.Add(http.CanonicalHeaderKey("x-apigateway-method"), req.Context.Method)
	httpRequest.Header.Add(http.CanonicalHeaderKey("x-apigateway-path"), req.Context.Path)
	httpRequest.Header.Add(http.CanonicalHeaderKey("x-apigateway-sourceip"), req.Context.SourceIP)
	httpRequest.Header.Add(http.CanonicalHeaderKey("x-forwarded-for"), req.Context.SourceIP)
	httpRequest.Header.Add(http.CanonicalHeaderKey("x-apigateway-stage"), req.Context.Stage)
	return httpRequest, nil
}

func addToHeader(req *http.Request, apiGwRequest events.APIGatewayRequest) (*http.Request, error) {
	apiGwContext, err := json.Marshal(apiGwRequest.Context)
	if err != nil {
		log.Println("Could not Marshal API GW context for custom header")
		return req, err
	}
	req.Header.Add(APIGwContextHeader, string(apiGwContext))
	return req, nil
}

func addToContext(ctx context.Context, req *http.Request, apiGwRequest events.APIGatewayRequest) *http.Request {
	lc, _ := functioncontext.FromContext(ctx)
	rc := requestContext{lambdaContext: lc, gatewayProxyContext: apiGwRequest.Context}
	ctx = context.WithValue(ctx, ctxKey{}, rc)
	return req.WithContext(ctx)
}

// GetAPIGatewayContextFromContext retrieve APIGatewayProxyRequestContext from context.Context
func GetAPIGatewayContextFromContext(ctx context.Context) (events.APIGatewayRequestContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContext)
	return v.gatewayProxyContext, ok
}

// GetRuntimeContextFromContext retrieve Lambda Runtime Context from context.Context
func GetRuntimeContextFromContext(ctx context.Context) (*functioncontext.FunctionContext, bool) {
	v, ok := ctx.Value(ctxKey{}).(requestContext)
	return v.lambdaContext, ok
}

type ctxKey struct{}

type requestContext struct {
	lambdaContext       *functioncontext.FunctionContext
	gatewayProxyContext events.APIGatewayRequestContext
}
