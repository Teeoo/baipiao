package jd

import (
	"fmt"
	. "github.com/teeoo/baipiao/cache"
	. "github.com/teeoo/baipiao/cron"
	json "github.com/tidwall/gjson"
	"log"
	"net/http"
	"net/url"
)

var ck *log.Logger

func init() {
	ck = initLogger("./logs/jd_refresh_ck", "刷新ck")
	_, err := Task.AddFunc("11 23 * * *", func() {
		ck = initLogger("./logs/jd_refresh_ck", "刷新ck")
		var data = Redis.Keys(ctx, "baipiao:ck:*")
		for _, s := range data.Val() {
			result := Redis.HGetAll(ctx, s)
			response, _ := client.R().SetCookies([]*http.Cookie{
				{
					Name:  "pin",
					Value: result.Val()["pt_pin"],
				}, {
					Name:  "wskey",
					Value: result.Val()["ws_key"],
				},
			}).
				SetHeader("Content-type", "application/x-www-form-urlencoded").
				SetHeader("User-Agent", "jdapp;iPhone;10.1.2;15.0;cc4a3fee7254710140e7ccc0443480e5d6b3ca68;network/wifi;model/iPhone12,1;addressid/2865568211;appBuild/167802;jdSupportDarkMode/0;Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148;supportJDSHWK/1").
				SetFormData(map[string]string{
					"area":           "24_2144_24463_58345",
					"body":           `{"to":"https:\/\/plogin.m.jd.com\/jd-mlogin\/static\/html\/appjmp_blank.html","action":"to"}`,
					"build":          `167814`,
					"client":         `apple`,
					"clientVersion":  "10.1.4",
					"d_brand":        "apple",
					"d_model":        "iPhone12,1",
					"eid":            "eidI46588123b9s3WNmzB1PXT6mwp+pzjqIUO1jZtHnWKawlX2malzeRTmEYkFhIKa7uFmZOsAplfKhmwEhP1pVEtTxxuF1WEejsCBMzayY7eJO0k/F5",
					"isBackground":   "N",
					"lang":           "zh_CN",
					"networkType":    "wifi",
					"networklibtype": "JDNetworkBaseAF",
					"openudid":       "cc4a3fee7254710140e7ccc0443480e5d6b3ca68",
					"osVersion":      "15.0",
					"partner":        "apple",
					"rfs":            "0000",
					"scope":          "01",
					"screen":         "828*1792",
					"sign":           "34f050268334cec22e6308cad6e30808",
					"st":             "1631514682231",
					"sv":             "100",
					"uemps":          "0-0",
					"uts":            "0f31TVRjBStbcw4pwwE+3b7KVET/7YtoREbp+NlxRkBC5ZeUb8Zb04tybQJjIV7ZxKMdVomksD/BxxSajUVwK+sGHGcwIYsuu9n9RWhn0lxdtTAVngsQTlc/uSZKp5kkIPVifDYLSb1j2/E3+s2dFXA/hrdJFsWzq8OUF+O9qyDdYm1A+/Lbc7gjiwNYI6KB1eqJaQbrI/43FwBpPY7DTQ==",
					"uuid":           "hjudwgohxzVu96krv/T6Hg==",
					"wifiBssid":      "9ffc61419a77e9326e7a0f1802b80dd7",
				}).
				Post("https://api.m.jd.com/client.action?functionId=genToken")
			_, _ = client.R().Post(fmt.Sprintf("%s?tokenKey=%s", json.Get(string(response.Body()), "url").String(), json.Get(string(response.Body()), "tokenKey").String()))
			u, _ := url.Parse("https://jd.com")
			for _, cookie := range client.GetClient().Jar.Cookies(u) {
				if cookie.Name == "pt_key" {
					ck.Printf("已刷新 %s ck", cookie.Name)
					Redis.HSet(ctx, s, "pt_key", cookie.Value)
				}
			}
		}
	})
	if err != nil {
		ck.Println(err)
	}
}
