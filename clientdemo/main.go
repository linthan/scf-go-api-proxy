package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/viper"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	scf "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/scf/v20180416"
	"github.com/tencentyun/scf-go-lib/events"
)

func main() {
	// 实例化一个认证对象，入参需要传入腾讯云账户 SecretId 和 SecretKey，此处还需注意密钥对的保密
	// 代码泄露可能会导致 SecretId 和 SecretKey 泄露，并威胁账号下所有资源的安全性。以下代码示例仅供参考，建议采用更安全的方式来使用密钥，请参见：https://cloud.tencent.com/document/product/1278/85305
	// 密钥可前往官网控制台 https://console.cloud.tencent.com/cam/capi 进行获取

	viper.SetConfigFile("./config.toml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("Can't read config:", err)
		os.Exit(1)
	}
	credential := common.NewCredential(
		viper.GetString("secretID"),
		viper.GetString("secretKey"),
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "scf.tencentcloudapi.com"
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, _ := scf.NewClient(credential, "ap-guangzhou", cpf)

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := scf.NewInvokeRequest()
	request.FunctionName = common.StringPtr("simple")
	request.InvocationType = common.StringPtr("RequestResponse") //# 接口参数,输入需要调用的函数名，RequestResponse(同步) 和 Event(异步)
	event := events.APIGatewayRequest{}
	event.Body = `{"a":1}`
	event.Method = "POST"
	event.Path = "/hello"
	event.Context.Method = "/POST"
	event.Context.Path = "/hello"
	eventb, _ := json.Marshal(event)
	request.ClientContext = common.StringPtr(string(eventb))

	// 返回的resp是一个InvokeResponse的实例，与请求对象对应
	response, err := client.Invoke(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		fmt.Printf("An API error has returned: %s", err)
		return
	}
	if err != nil {
		panic(err)
	}
	// 输出json格式的字符串回包
	fmt.Printf("%s", response.ToJsonString())
}
