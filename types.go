package weiyunsdkgo

import (
	"encoding/json"
	"errors"
	"fmt"
)

var (
	ErrCode403 = errors.New("http code 403")
)

type RespHeader struct {
	Retcode int    `json:"retcode"`
	Retmsg  string `json:"retmsg"`

	Cmd     int    `json:"cmd"`
	CmdName string `json:"cmdName"`

	Uin  int64 `json:"uin"`
	Type int   `json:"type"`

	// flag
	ZipFlag int `json:"zip_flag"`
	Encrypt int `json:"encrypt"`

	Seq int64 `json:"seq"`
}

type Resp struct {
	Code int    `json:"ret"`
	Msg  string `json:"msg"`

	Data struct {
		RspHeader RespHeader `json:"rsp_header"`
		RspBody   struct {
			RspMsgBody json.RawMessage `json:"RspMsg_body"`
		} `json:"rsp_body"`
	} `json:"data"`
}

func (r *Resp) HasError() bool {
	return r.Data.RspHeader.Retcode != 0 || r.Code != 0
}

func (r *Resp) Error() string {
	if r.Msg != "" {
		return fmt.Sprintf("errcode:%d, errmsg:%s", r.Code, r.Msg)
	}
	return fmt.Sprintf("cmd:%d, cmdName:%s, seq:%d, errcode:%d, errmsg:%s",
		r.Data.RspHeader.Cmd,
		r.Data.RspHeader.CmdName,
		r.Data.RspHeader.Seq,
		r.Data.RspHeader.Retcode,
		r.Data.RspHeader.Retmsg)
}

func (r *Resp) GetHeader() RespHeader {
	return r.Data.RspHeader
}

func (r *Resp) GetBody() json.RawMessage {
	return r.Data.RspBody.RspMsgBody
}
