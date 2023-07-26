package weiyunsdkgo

import (
	"net/http"
	"net/url"

	"github.com/go-resty/resty/v2"
)

var baseUrl, _ = url.Parse("https://www.weiyun.com")

type RestyOption func(request *resty.Request)
type Json map[string]any

type WeiYunClient struct {
	Client *resty.Client
	flag   int32

	token *OpenApiToken
	flag2 int32
}

func NewWeiYunClient() *WeiYunClient {
	return &WeiYunClient{
		Client: resty.New(),
	}
}

func (c *WeiYunClient) SetClient(client *http.Client) *WeiYunClient {
	c.Client = resty.NewWithClient(client)
	return c
}

func (c *WeiYunClient) SetRestyClient(client *resty.Client) *WeiYunClient {
	c.Client = client
	return c
}

func (c *WeiYunClient) SetProxy(proxy string) *WeiYunClient {
	c.Client.SetProxy(proxy)
	return c
}

// 设置登录Cookie
func (c *WeiYunClient) SetCookies(cks []*http.Cookie) *WeiYunClient {
	for _, ck := range cks {
		ck.Domain = ".weiyun.com"
		ck.Path = "/"
	}
	c.GetCookieJar().SetCookies(baseUrl, cks)
	return c
}

// 设置登录Cookie字符串
func (c *WeiYunClient) SetCookiesStr(str string) *WeiYunClient {
	return c.SetCookies(ParseCookieStr(str))
}

// 获取登录Cookie
// 内部未拷贝，谨慎修改
func (c *WeiYunClient) GetCookies() []*http.Cookie {
	return c.GetCookieJar().Cookies(baseUrl)
}

func (c *WeiYunClient) GetCookieJar() http.CookieJar {
	return c.Client.GetClient().Jar
}

// 设置OpenApi Token
func (c *WeiYunClient) SetOpenApiToken(token *OpenApiToken) *WeiYunClient {
	c.token = token
	return c
}

// 获取OpenApi Token
func (c *WeiYunClient) GetOpenApiToken() *OpenApiToken {
	return c.token
}
