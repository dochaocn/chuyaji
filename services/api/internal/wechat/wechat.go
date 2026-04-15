package wechat

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Code2SessionResp struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

func Code2Session(appID, appSecret, code string) (*Code2SessionResp, error) {
	u := url.Values{}
	u.Set("appid", appID)
	u.Set("secret", appSecret)
	u.Set("js_code", code)
	u.Set("grant_type", "authorization_code")
	api := "https://api.weixin.qq.com/sns/jscode2session?" + u.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(api)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out Code2SessionResp
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if out.ErrCode != 0 {
		return nil, fmt.Errorf("wechat: %d %s", out.ErrCode, out.ErrMsg)
	}
	if out.OpenID == "" {
		return nil, fmt.Errorf("wechat: empty openid")
	}
	return &out, nil
}
