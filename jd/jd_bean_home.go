package jd

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	. "github.com/teeoo/baipiao/cache"
	. "github.com/teeoo/baipiao/http"
	"github.com/teeoo/baipiao/typefac"
	json "github.com/tidwall/gjson"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type BeanHome struct{}

var (
	beanLogger *log.Logger
)

func init() {
	typefac.RegisterType(BeanHome{})
	log.Println("京东APP->我的->签到领京豆->领额外奖励")
}

// Run @Cron 45 0 * * *
func (c BeanHome) Run() {
	beanLogger = initLogger("./logs/jd_amusement", "京小鸽游乐寄")
	var data = Redis.Keys(ctx, "baipiao:ck:*")
	go func() {
		for _, s := range data.Val() {
			result := Redis.HGetAll(ctx, s)
			HttpClient.SetDebug(false).SetCookies([]*http.Cookie{
				{
					Name:  "pt_pin",
					Value: result.Val()["pt_pin"],
				}, {
					Name:  "pt_key",
					Value: result.Val()["pt_key"],
				},
			})
			beanTasks(HttpClient.R(), result.Val()["pt_pin"])
			award(HttpClient.R(), "home", result.Val()["pt_pin"])
			beanGoodsTasks(HttpClient.R(), result.Val()["pt_pin"])
			award(HttpClient.R(), "feeds", result.Val()["pt_pin"])
		}
	}()
}

func _bean(c *resty.Request, functionId, body string) string {
	params := url.Values{}
	params.Add("functionId", functionId)
	params.Add("appid", "ld")
	params.Add("clientVersion", "10.0.11")
	params.Add("client", "apple")
	params.Add("eu", fp())
	params.Add("fv", fp())
	params.Add("osVersion", "11")
	params.Add("uuid", fmt.Sprintf("%s%s", fp(), fp()))
	params.Add("openudid", fmt.Sprintf("%s%s", fp(), fp()))
	params.Add("body", body)
	u := fmt.Sprintf("https://api.m.jd.com/client.action?%s", params.Encode())
	timer := time.NewTimer(1 * time.Second)
	select {
	case <-timer.C:
		resp, err := c.SetHeaders(map[string]string{
			//"content-type": "application/x-www-form-urlencoded",
			"user-agent": UserAgent(),
			"referer":    "https://h5.m.jd.com/rn/2E9A2bEeqQqBP9juVgPJvQQq6fJ/index.html",
		}).SetBody(body).Post(u)
		if err != nil {
			log.Println(err)
		}
		timer.Stop()
		return string(resp.Body())
	}
}

func award(c *resty.Request, source, user string) {
	resp := _bean(c, "beanHomeTask", fmt.Sprintf(`{"source":"%s","awardFlag":true}`, source))
	if json.Get(resp, "code").String() != "0" || json.Get(resp, "errorCode").String() != "" {
		beanLogger.Printf("%s 领取京豆奖励失败: %s", user, json.Get(resp, "errorMessage").String())
	} else {
		beanLogger.Printf("%s 领取京豆奖励, 获得京豆: %s", user, json.Get(resp, "data.beanNum").String())
	}
}

func beanTasks(c *resty.Request, user string) {
	resp := _bean(c, "findBeanHome", fmt.Sprintf(`{"source": "wojing2", "orderId": "null", "rnVersion": "3.9", "rnClient": "1"}`))
	if json.Get(resp, "code").String() != "0" {
		beanLogger.Printf("%s 获取首页数据失败", user)
	}
	if json.Get(resp, "data.taskProgress").Int() == json.Get(resp, "data.taskThreshold").Int() {
		beanLogger.Printf("%s 今日已完成领额外京豆任务", user)
	}
	for i := range [6]int{} {
		body := fmt.Sprintf(`{"type": "%s", "source": "home", "awardFlag": false, "itemId": ""}`, strconv.Itoa(i+1))
		timer := time.NewTimer(1 * time.Second)
		select {
		case <-timer.C:
			result := _bean(c, "beanHomeTask", body)
			if json.Get(result, "errorCode").String() != "" {
				beanLogger.Printf("%s 领额外京豆任务失败 %s", user, json.Get(result, "errorMessage").String())
			} else {
				beanLogger.Printf("%s 领额外京豆任务进度:%s/%s", user, json.Get(result, "data.taskProgress").String(), json.Get(result, "data.taskThreshold").String())
			}
		}
		timer.Stop()
	}
}

func beanGoodsTasks(c *resty.Request, user string) {
	resp := _bean(c, "homeFeedsList", fmt.Sprintf(`{"page": 1}`))
	if json.Get(resp, "code").String() != "0" || json.Get(resp, "errorCode").String() != "" {
		beanLogger.Printf("%s 浏览商品任务 %s", user, resp)
	}
	if json.Get(resp, "data.taskProgress").String() == json.Get(resp, "data.taskThreshold").String() {
		beanLogger.Printf("%s 今日已完成浏览商品任务", user)
	}
	for range [3]int{} {
		timer := time.NewTimer(1 * time.Second)
		select {
		case <-timer.C:
			result := _bean(c, "beanHomeTask", fmt.Sprintf(`{"type": "1", "skuId": "%s", "awardFlag": false, "source": "feeds","scanTime":"%s"}`, strconv.Itoa(randomInt(10000000, 20000000)), strconv.Itoa(int(time.Now().Unix()*1000))))
			if json.Get(result, "errorCode").String() != "" {
				beanLogger.Printf("%s 浏览商品任务 %s", user, json.Get(result, "errorMessage").String())
			} else {
				beanLogger.Printf("%s 浏览商品任务:%s/%s", user, json.Get(result, "data.taskProgress").String(), json.Get(result, "data.taskThreshold").String())
			}
		}
		timer.Stop()
	}
}
