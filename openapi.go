package weiyunsdkgo

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-resty/resty/v2"
)

const (
	ClientID     = "iciY14xT8lJF0SrlfvIhuTD36VeOkm9R"
	ClientSecret = "26gJq9sDVHHxVUrAJy0COQ196v0EVXaW"
)

func (c *WeiYunClient) OpenApiRequest(method string, url string, data Json, resp any, opts ...RestyOption) (*resty.Response, error) {
	if c.token == nil {
		return nil, ErrTokenIsNil
	}

	resp_, err := c.openApiRequest(method, url, data, resp, opts...)

	if err != nil && (strings.Contains(err.Error(), "令牌") || strings.Contains(err.Error(), "Token")) {
		if strings.Contains(err.Error(), "失效") || strings.Contains(err.Error(), "过期") {
			if atomic.CompareAndSwapInt32(&c.flag2, 0, 1) {
				_, err2 := c.OpenApiRefreshToken(c.token.RefreshToken)
				atomic.SwapInt32(&c.flag2, 0)
				if err2 != nil {
					return resp_, errors.Join(err, err2, ErrTokenExpiration)
				}
			}
			for atomic.LoadInt32(&c.flag2) != 0 {
				time.Sleep(time.Second)
			}
			resp_, err = c.openApiRequest(method, url, data, resp, opts...)
		}
	}
	return resp_, err
}

func (c *WeiYunClient) openApiRequest(method string, url string, data Json, resp any, opts ...RestyOption) (*resty.Response, error) {
	var erron OpenApiErron
	req := c.Client.R().SetAuthScheme(c.token.TokenType).SetAuthToken(c.token.AccessToken)

	if resp != nil {
		req.SetResult(resp)
	}

	if data != nil {
		req.SetBody(data)
	}

	for _, opt := range opts {
		opt(req)
	}

	resp_, err := req.SetError(&erron).Execute(method, url)
	if err != nil {
		return nil, err
	}
	if erron.IsError() {
		return resp_, &erron
	}
	return resp_, nil
}

//func (c *WeiYunClient)

// 获取 OpenApi Token
// 请先登录或设置Cookie
// 当前仅支持QQ
func (c *WeiYunClient) OpenApiGetToken() (*OpenApiToken, error) {
	var code string

	// 获取code
	var erron OpenApiErron
	_, err := c.Client.R().
		SetError(&erron).
		SetQueryParams(map[string]string{
			"client_id":     ClientID,
			"redirect_uri":  "wps-office-android://www.wps.cn:12345",
			"response_type": "code",
			"state":         "wps_weiyun_login_state",
			"login_type":    "4",
		}).
		Get("https://user.weiyun.com/twoa/v1/auth/authorize")

	if erron.IsError() {
		return nil, &erron
	}

	if err != nil {
		codes := regexp.MustCompile(`wps-office-android://www.wps.cn:12345\?code=([^"]+)`).FindStringSubmatch(err.Error())
		if len(codes) == 0 {
			return nil, err
		}
		code = codes[1]
	}

	// 获取token
	var token OpenApiToken
	_, err = c.Client.R().
		SetError(&erron).
		SetResult(&token).
		SetQueryParams(map[string]string{
			"grant_type":    "authorization_code",
			"client_id":     ClientID,
			"client_secret": ClientSecret,
			"code":          code,
		}).
		Get("https://user.weiyun.com/twoa/v1/auth/token")
	if err != nil {
		return nil, err
	}

	c.SetOpenApiToken(&token)
	return &token, nil
}

// 刷新OpenApi Token
func (c *WeiYunClient) OpenApiRefreshToken(refreshToken string) (*OpenApiToken, error) {
	var token OpenApiToken
	var erron OpenApiErron

	_, err := c.Client.R().
		SetError(&erron).
		SetResult(&token).
		SetQueryParams(map[string]string{
			"client_id":     ClientID,
			"client_secret": ClientSecret,
			"refresh_token": refreshToken,
		}).
		Get("https://user.weiyun.com/twoa/v1/auth/refresh_token")
	if err != nil {
		return nil, err
	}

	c.SetOpenApiToken(&token)
	return &token, nil
}

type OpenApiGetUserInfoData struct {
	DefaultDir string `json:"default_dir"`
	Uin        int    `json:"uin"`
	NickName   string `json:"nick_name"`
	UsedSpace  int    `json:"used_space"`
	IsVip      bool   `json:"is_vip"`
	TotalSpace int64  `json:"total_space"`
	WeiyunDir  string `json:"weiyun_dir"`
	VipLevel   int    `json:"vip_level"`
}

func (o *OpenApiGetUserInfoData) GetRootDirID() string {
	return GetDirFileIDFormUrl(o.WeiyunDir)
}

func (c *WeiYunClient) OpenApiGetUserInfo(opts ...RestyOption) (*OpenApiGetUserInfoData, error) {
	var resp OpenApiGetUserInfoData
	_, err := c.OpenApiRequest(http.MethodGet, "https://user.weiyun.com/twoa/v1/users/get_info", nil, &resp, opts...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

type OpenApiFolder struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

func (o *OpenApiFolder) GetFolderID() string {
	return GetDirFileIDFormUrl(o.URL)
}

type OpenApiFile struct {
	Name  string `json:"name"`
	URL   string `json:"url"`
	Size  int    `json:"size"`
	Ctime int64  `json:"ctime"`
	Mtime int64  `json:"mtime"`
	Sha1  string `json:"sha"`
}

func (o *OpenApiFile) GetFileID() string {
	return GetDirFileIDFormUrl(o.URL)
}

type OpenApiGetFileListData struct {
	Finished       bool            `json:"finished"`
	TotalDirCount  int             `json:"total_dir_count"`
	TotalFileCount int             `json:"total_file_count"`
	DirList        []OpenApiFolder `json:"dir_list"`
	FileList       []OpenApiFile   `json:"file_list"`
}

func (c *WeiYunClient) OpenApiGetFileList(dirID string, offset int64, count int, opts ...RestyOption) (*OpenApiGetFileListData, error) {
	var resp OpenApiGetFileListData
	_, err := c.OpenApiRequest(http.MethodGet, "https://user.weiyun.com/twoa/v1/dirs/{rs_id}/list", nil, &resp, append([]RestyOption{
		func(request *resty.Request) {
			request.SetPathParam("rs_id", dirID)
			request.SetQueryParams(map[string]string{
				"offset": fmt.Sprint(offset),
				"count":  fmt.Sprint(count),
			})
		},
	}, opts...)...)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

func (c *WeiYunClient) OpenApiDownloadFile(fileID string, strat, end int64, opts ...RestyOption) (io.ReadCloser, error) {
	resp, err := c.OpenApiRequest(http.MethodGet, "https://user.weiyun.com/twoa/v1/files/{rs_id}/download", nil, nil, append([]RestyOption{
		func(request *resty.Request) {
			request.SetDoNotParseResponse(true).
				SetHeader("Range", fmt.Sprintf("bytes=%d-%d", strat, end)).
				SetPathParam("rs_id", fileID)
		}}, opts...)...)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse.Body, nil
}
