package jd

import (
	"context"
	j "encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	. "github.com/teeoo/baipiao/cache"
	"github.com/teeoo/baipiao/typefac"
	json "github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type Sign struct{}

var (
	ctx    = context.Background()
	client *resty.Client
)

type SignInfo struct {
	EnActK       string `json:"enActK"`
	IsFloatLayer bool   `json:"isFloatLayer"`
	RuleSrv      string `json:"ruleSrv"`
	SignID       string `json:"signId"`
}

func init() {
	PathExists("./logs/jd_sign")
	typefac.RegisterType(Sign{})
	log.Println("京东签到合集")
}

// Run @Cron 0 3,19 * * *
func (c Sign) Run() {
	loggerFile, err := os.OpenFile(fmt.Sprintf("%s/%s.log", "./logs/jd_sign", time.Now().Format("2006-01-02-15-04-05")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, loggerFile))
	log.SetPrefix(fmt.Sprintf("[%s]", "京东签到合集"))
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile | log.Lshortfile)
	var data = Redis.Keys(ctx, "baipiao:ck:*")
	go func() {
		for _, s := range data.Val() {
			result := Redis.HGetAll(ctx, s)
			client = resty.New().SetDebug(false).SetCookies([]*http.Cookie{
				{
					Name:  "pt_pin",
					Value: result.Val()["pt_pin"],
				}, {
					Name:  "pt_key",
					Value: result.Val()["pt_key"],
				},
			}).SetHeader("User-Agent", "jdapp;iPhone;10.1.2;15.0;cc4a3fee7254710140e7ccc0443480e5d6b3ca68;network/wifi;model/iPhone12,1;addressid/2865568211;appBuild/167802;jdSupportDarkMode/0;Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148;supportJDSHWK/1")
			beanSign(client.R(), result.Val()["pt_pin"])
			jdShopWomen(client.R(), result.Val()["pt_pin"])
			jdShopCard(client.R(), result.Val()["pt_pin"])
			jdShopBook(client.R(), result.Val()["pt_pin"])
			jdShopAccompany(client.R(), result.Val()["pt_pin"])
			jdShopSuitcase(client.R(), result.Val()["pt_pin"])
			jdShopShoes(client.R(), result.Val()["pt_pin"])
			jdShopFoodMarket(client.R(), result.Val()["pt_pin"])
			jdShopClothing(client.R(), result.Val()["pt_pin"])
			jdShopHealth(client.R(), result.Val()["pt_pin"])
			jdShopSecondHand(client.R(), result.Val()["pt_pin"])
			jdShopSchool(client.R(), result.Val()["pt_pin"])
		}
	}()
}

// 签到领京豆
func beanSign(c *resty.Request, user string) {
	resp, _ := c.Post("https://api.m.jd.com/client.action?functionId=signBeanIndex&appid=ld")
	status := json.Get(string(resp.Body()), "data.status").String()
	if status == "1" {
		log.Println(user, "签到京豆签到成功!")
	} else if status == "2" {
		log.Println(user, "签到京豆今日已签到")
	} else {
		log.Println(user, "签到京豆签到失败")
	}
}

// 京东商城签到
func jdShop(c *resty.Request, name, data, user string) {
	resp, _ := c.SetHeader("Content-type", "application/x-www-form-urlencoded").SetFormData(map[string]string{
		"body": data,
	}).Post("https://api.m.jd.com/?client=wh5&functionId=qryH5BabelFloors")
	body := string(resp.Body())
	floatLayerText := json.Get(body, "floatLayerList.#.params").Array()
	for _, result := range floatLayerText {
		if json.Get(result.String(), "enActK").String() != "" {
			params, _ := j.Marshal(result.String())
			jdShopSign(c, name, string(params), user)
		}
	}
	floorList := json.Get(body, "floorList").Array()
	for _, result := range floorList {
		if result.Map()["template"].String() == "signIn" {
			signInfo := result.Map()["signInfos"]
			if signInfo.Map()["signStat"].String() == "1" {
				log.Printf("%s,%s今日已签到!", name, user)
			} else {
				//params := new(SignInfo)
				params, _ := j.Marshal(signInfo.Map()["params"].String())
				jdShopSign(c, name, string(params), user)
			}
		}
	}
}

func jdShopSign(c *resty.Request, name, body, user string) {
	resp, err := c.SetHeader("Content-type", "application/x-www-form-urlencoded").SetFormData(map[string]string{
		"body":   fmt.Sprintf(`{"params":%s}`, body),
		"client": "wh5",
	}).Post("https://api.m.jd.com/client.action?functionId=userSign")
	if err != nil {
		log.Println(name, user, "签到异常", err)
	}
	if strings.Contains(string(resp.Body()), "签到成功") {
		log.Println(name, user, "签到成功")
	} else if strings.Contains(string(resp.Body()), "已签到") {
		log.Println(name, user, "今日已签到")
	} else {
		log.Println(name, user, "签到失败")
	}
}

func jdShopWomen(c *resty.Request, user string) {
	jdShop(c, "京东商城-女装", `{"activityId":"DpSh7ma8JV7QAxSE2gJNro8Q2h9"}`, user)
}

func jdShopCard(c *resty.Request, user string) {
	jdShop(c, "京东商城-女装", `{"activityId":"7e5fRnma6RBATV9wNrGXJwihzcD"}`, user)
}

func jdShopBook(c *resty.Request, user string) {
	jdShop(c, "京东商城-图书", `{"activityId":"3SC6rw5iBg66qrXPGmZMqFDwcyXi"}`, user)
}

func jdShopAccompany(c *resty.Request, user string) {
	jdShop(c, "京东商城-陪伴", `{"activityId":"kPM3Xedz1PBiGQjY4ZYGmeVvrts"}`, user)
}

func jdShopSuitcase(c *resty.Request, user string) {
	jdShop(c, "京东商城-箱包", `{"activityId":"ZrH7gGAcEkY2gH8wXqyAPoQgk6t"}`, user)
}

func jdShopShoes(c *resty.Request, user string) {
	jdShop(c, "京东商城-鞋靴", `{"activityId":"4RXyb1W4Y986LJW8ToqMK14BdTD"}`, user)
}

func jdShopFoodMarket(c *resty.Request, user string) {
	jdShop(c, "京东商城-菜场", `{"activityId":"Wcu2LVCFMkBP3HraRvb7pgSpt64"}`, user)
}

func jdShopClothing(c *resty.Request, user string) {
	jdShop(c, "京东商城-服饰", `{"activityId":"4RBT3H9jmgYg1k2kBnHF8NAHm7m8"}`, user)
}

func jdShopHealth(c *resty.Request, user string) {
	jdShop(c, "京东商城-健康", `{"activityId":"w2oeK5yLdHqHvwef7SMMy4PL8LF"}`, user)
}

func jdShopSecondHand(c *resty.Request, user string) {
	jdShop(c, "京东拍拍-二手", `{"activityId":"3S28janPLYmtFxypu37AYAGgivfp"}`, user)
}

func jdShopSchool(c *resty.Request, user string) {
	jdShop(c, "京东商城-校园", `{"activityId":"2QUxWHx5BSCNtnBDjtt5gZTq7zdZ"}`, user)
}
