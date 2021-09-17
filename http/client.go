package http

import (
	"github.com/go-resty/resty/v2"
)

var (
	HttpClient *resty.Client
)

func init() {
	HttpClient = resty.New().SetDebug(false).SetHeader("User-Agent", "jdapp;iPhone;10.1.2;15.0;cc4a3fee7254710140e7ccc0443480e5d6b3ca68;network/wifi;model/iPhone12,1;addressid/2865568211;appBuild/167802;jdSupportDarkMode/0;Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148;supportJDSHWK/1")
}