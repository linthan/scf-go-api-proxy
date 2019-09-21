package core_test

import (
	"context"
	"crypto/rand"
	"os"

	"github.com/linthan/scf-go-api-proxy/core"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/tencentyun/scf-go-lib/events"
)

var _ = Describe("RequestAccessor tests", func() {
	Context("event conversion", func() {
		accessor := core.RequestAccessor{}
		basicRequest := getProxyRequest("/hello", "GET")
		It("Correctly converts a basic event", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basicRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))
		})
		basicRequest = getProxyRequest("/hello", "get")
		It("Converts method to uppercase", func() {
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(basicRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))
		})

		binaryBody := make([]byte, 256)
		_, err := rand.Read(binaryBody)
		if err != nil {
			Fail("Could not generate random binary body")
		}

		mqsRequest := getProxyRequest("/hello", "GET")
		mqsRequest.QueryString = map[string][]string{
			"hello": []string{"1"},
			"world": []string{"2"},
		}
		It("Populates multiple value query string correctly", func() {
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), mqsRequest)
			Expect(err).To(BeNil())
			Expect("/hello").To(Equal(httpReq.URL.Path))
			Expect("GET").To(Equal(httpReq.Method))

			query := httpReq.URL.Query()
			Expect(2).To(Equal(len(query)))
			Expect(query["hello"]).ToNot(BeNil())
			Expect(query["world"]).ToNot(BeNil())
			Expect(1).To(Equal(len(query["hello"])))
			Expect(1).To(Equal(len(query["world"])))
			Expect("1").To(Equal(query["hello"][0]))
			Expect("2").To(Equal(query["world"][0]))
		})

	})

	Context("StripBasePath tests", func() {
		accessor := core.RequestAccessor{}
		It("Adds prefix slash", func() {
			basePath := accessor.StripBasePath("app1")
			Expect("/app1").To(Equal(basePath))
		})

		It("Removes trailing slash", func() {
			basePath := accessor.StripBasePath("/app1/")
			Expect("/app1").To(Equal(basePath))
		})

		It("Ignores blank strings", func() {
			basePath := accessor.StripBasePath("  ")
			Expect("").To(Equal(basePath))
		})
	})

	Context("Retrieves API Gateway context", func() {
		It("Returns a correctly unmarshalled object", func() {
			contextRequest := getProxyRequest("orders", "GET")
			contextRequest.Context = getRequestContext()

			accessor := core.RequestAccessor{}
			// calling old method to verify reverse compatibility
			httpReq, err := accessor.ProxyEventToHTTPRequest(contextRequest)
			Expect(err).To(BeNil())
			headerContext, err := accessor.GetAPIGatewayContext(httpReq)
			Expect(err).To(BeNil())
			Expect(headerContext).ToNot(BeNil())
			Expect("x").To(Equal(headerContext.ServiceID))
			Expect("x").To(Equal(headerContext.RequestID))
			proxyContext, ok := core.GetAPIGatewayContextFromContext(httpReq.Context())
			// should fail because using header proxy method
			Expect(ok).To(BeFalse())

			httpReq, err = accessor.EventToRequestWithContext(context.Background(), contextRequest)
			Expect(err).To(BeNil())
			proxyContext, ok = core.GetAPIGatewayContextFromContext(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect("x").To(Equal(proxyContext.ServiceID))
			Expect("x").To(Equal(proxyContext.RequestID))
			Expect("prod").To(Equal(proxyContext.Stage))
			runtimeContext, ok := core.GetRuntimeContextFromContext(httpReq.Context())
			Expect(ok).To(BeTrue())
			Expect(runtimeContext).To(BeNil())

		})

		It("Populates the default hostname correctly", func() {
			basicRequest := getProxyRequest("orders", "GET")
			accessor := core.RequestAccessor{}
			httpReq, err := accessor.ProxyEventToHTTPRequest(basicRequest)
			Expect(err).To(BeNil())
			Expect(core.DefaultServerAddress).To(Equal("https://" + httpReq.Host))
			Expect(core.DefaultServerAddress).To(Equal("https://" + httpReq.URL.Host))
		})

		It("Uses a custom hostname", func() {
			myCustomHost := "http://my-custom-host.com"
			os.Setenv(core.CustomHostVariable, myCustomHost)
			basicRequest := getProxyRequest("orders", "GET")
			accessor := core.RequestAccessor{}
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basicRequest)
			Expect(err).To(BeNil())
			Expect(myCustomHost).To(Equal("http://" + httpReq.Host))
			Expect(myCustomHost).To(Equal("http://" + httpReq.URL.Host))
			os.Unsetenv(core.CustomHostVariable)
		})

		It("Strips terminating / from hostname", func() {
			myCustomHost := "http://my-custom-host.com"
			os.Setenv(core.CustomHostVariable, myCustomHost+"/")
			basicRequest := getProxyRequest("orders", "GET")
			accessor := core.RequestAccessor{}
			httpReq, err := accessor.EventToRequestWithContext(context.Background(), basicRequest)
			Expect(err).To(BeNil())
			Expect(myCustomHost).To(Equal("http://" + httpReq.Host))
			Expect(myCustomHost).To(Equal("http://" + httpReq.URL.Host))
			os.Unsetenv(core.CustomHostVariable)
		})
	})

})

func getProxyRequest(path string, method string) events.APIGatewayRequest {
	return events.APIGatewayRequest{
		Path:   path,
		Method: method,
	}
}

func getRequestContext() events.APIGatewayRequestContext {
	return events.APIGatewayRequestContext{
		ServiceID: "x",
		RequestID: "x",
		Stage:     "prod",
	}
}
