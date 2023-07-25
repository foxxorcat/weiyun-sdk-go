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

func (c *WeiYunClient) SetCookies(cks []*http.Cookie) *WeiYunClient {
	c.GetCookieJar().SetCookies(baseUrl, cks)
	return c
}

func (c *WeiYunClient) SetCookiesStr(str string) *WeiYunClient {
	return c.SetCookies(ParseCookie(str))
}

func (c *WeiYunClient) GetCookies() []*http.Cookie {
	return c.GetCookieJar().Cookies(baseUrl)
}

func (c *WeiYunClient) GetCookieJar() http.CookieJar {
	return c.Client.GetClient().Jar
}
