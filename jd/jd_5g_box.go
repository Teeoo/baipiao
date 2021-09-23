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

type Box5G struct{}

var (
	boxLogger *log.Logger
)

func init() {
	typefac.RegisterType(Box5G{})
	log.Println("京东APP->营业厅->领京豆, 5G盲盒做任务抽奖")
}

// Run 8 0 * * *
func (c Box5G) Run() {
	boxLogger = initLogger("./logs/jd_5g_box", "5G盲盒")
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
			coin(HttpClient.R(), result.Val()["pt_pin"])
			shareCode(HttpClient.R(), result.Val()["pt_pin"])
			lottery(HttpClient.R(), result.Val()["pt_pin"])
			taskList(HttpClient.R(), result.Val()["pt_pin"], 11)
			browseGoods(HttpClient.R(), result.Val()["pt_pin"])
			lottery(HttpClient.R(), result.Val()["pt_pin"])
		}
	}()
}

func _box(c *resty.Request, body string) string {
	timestamp := time.Now().Unix()
	params := url.Values{}
	params.Add("appid", "blind-box")
	params.Add("functionId", "blindbox_prod")
	params.Add("body", body)
	params.Add("t", strconv.FormatInt(timestamp, 10))
	params.Add("loginType", "2")
	//var mapBody map[string]string
	//_ = j.Unmarshal([]byte(body), &mapBody)
	u := fmt.Sprintf("https://api.m.jd.com/api?%s", params.Encode())
	timer := time.NewTimer(1 * time.Second)
	select {
	case <-timer.C:
		resp, err := c.SetHeaders(map[string]string{
			"content-type": "application/x-www-form-urlencoded",
			"user-agent":   USER_AGENT(),
			"referer":      "https://blindbox5g.jd.com",
		}).Post(u)
		if err != nil {
			log.Println(err)
		}
		timer.Stop()
		return string(resp.Body())
	}
}

// 收信号值
func coin(c *resty.Request, user string) {
	data := _box(c, fmt.Sprintf(`{"apiMapping": "/active/getCoin"}`))
	if json.Get(data, "code").Int() == 200 {
		boxLogger.Printf("%s 成功收取信号值 %s", user, json.Get(data, "data").String())
	} else {
		boxLogger.Printf("%s 收取信号值失败 %s", user, json.Get(data, "msg").String())
	}
}

// 获取助力码
func shareCode(c *resty.Request, user string) {
	data := _box(c, fmt.Sprintf(`{"apiMapping": "/active/shareUrl"}`))
	if json.Get(data, "code").Int() == 200 {
		boxLogger.Printf("%s 助力码 %s", user, json.Get(data, "data").String())
	} else {
		boxLogger.Printf("%s 获取助力码失败 %s", user, json.Get(data, "msg").String())
	}
}

// 做任务
func taskList(c *resty.Request, user string, max int) {
	data := _box(c, fmt.Sprintf(`{"apiMapping": "/active/taskList"}`))
	if json.Get(data, "code").Int() == 200 {
		task := json.Get(data, "data")
		//boxLogger.Printf("%s %s", user, task.Map())
		var flag = false
		for _, result := range task.Map() {
			if result.Map()["type"].Int() != 5 || result.Map()["type"].Int() != 8 {
				if result.Map()["type"].Int() == 4 {
					flag = true
					_box(c, fmt.Sprintf(`{"skuId": "%s", "apiMapping": "/active/browseProduct"}`, result.Map()["skuId"].String()))
					boxLogger.Printf("%s 浏览商品 %s", user, result.Map()["skuId"].String())
				}

				if result.Map()["type"].Int() == 2 {
					flag = true
					_box(c, fmt.Sprintf(`{"shopId": "%s", "apiMapping": "/active/followShop"}`, result.Map()["shopId"].String()))
					boxLogger.Printf("%s 关注店铺 %s ,获得信号值 %s %s", user, result.Map()["shopId"].String(), result.Map()["coinNum"], result)
				}

				if result.Map()["type"].Int() == 1 {
					flag = true
					_box(c, fmt.Sprintf(`{"activeId": "%s", "apiMapping": "/active/strollActive"}`, result.Map()["activeId"].String()))
					boxLogger.Printf("%s 浏览会场 %s", user, result.Map()["skuId"].String())
				}

			}
		}
		if flag {
			boxLogger.Printf("%s 8秒后领取任务奖励!", user)
			timer := time.NewTimer(8 * time.Second)
			select {
			case <-timer.C:
				for _, result := range task.Map() {
					if result.Map()["type"].Int() != 5 || result.Map()["type"].Int() != 8 || result.Map()["type"].Int() != 2 {
						resp := _box(c, fmt.Sprintf(`{"type": "%s", "apiMapping": "/active/taskCoin"}`, result.Map()["type"].String()))
						if json.Get(resp, "code").Int() == 200 {
							boxLogger.Printf("%s 成功领取任务奖励,获得信号值 %s 京豆: %s", user, json.Get(resp, "data.coinNum"), json.Get(resp, "data.jbeanNum"))
						} else {
							boxLogger.Printf("%s 无法领取任务奖励 %s", user, json.Get(resp, "msg"))
						}
					}
				}
			}
			timer.Stop()
		}
		go taskList(c, user, max-1)
	} else {
		boxLogger.Printf("%s 获取任务列表失败 %s", user, json.Get(data, "msg").String())
	}
}

// 抽奖
func lottery(c *resty.Request, user string) {
	data := _box(c, fmt.Sprintf(`{"apiMapping": "/prize/lottery"}`))
	if json.Get(data, "code").Int() == 200 {
		boxLogger.Printf("%s 奖品 %s", user, json.Get(data, "data").String())
	} else {
		boxLogger.Printf("%s 抽奖失败 %s", user, json.Get(data, "msg").String())
	}
}

// 浏览精彩好物
func browseGoods(c *resty.Request, user string) {
	data := _box(c, fmt.Sprintf(`{"apiMapping": "/active/conf"}`))
	if json.Get(data, "code").Int() == 200 {
		itemList := json.Get(data, "data.skuList").Array()
		for _, result := range itemList {
			if result.Map()["state"].Int() != 2 {
				timer := time.NewTimer(2 * time.Second)
				select {
				case <-timer.C:
					_box(c, fmt.Sprintf(`{"type": "0", "id": "%s", "apiMapping": "/active/homeGoBrowse"}`, result.Map()["id"].String()))
					resp := _box(c, fmt.Sprintf(`{"type": "0", "id": "%s", "apiMapping": "/active/taskHomeCoin"}`, result.Map()["id"].String()))
					if json.Get(resp, "code").Int() == 200 {
						boxLogger.Printf("%s 浏览商品 %s", user, result.Map()["item"].String(), resp)
					}
				}
				timer.Stop()
			}
		}
	} else {
		boxLogger.Printf("%s 获取精彩好物列表失败 %s", user, json.Get(data, "msg").String())
	}
}
