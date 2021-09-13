package jd

import (
	"context"
	"github.com/go-resty/resty/v2"
	. "github.com/teeoo/baipiao/cache"
	"github.com/teeoo/baipiao/typefac"
	json "github.com/tidwall/gjson"
	"log"
	"net/http"
)

type Sign struct{}

var (
	ctx    = context.Background()
	client *resty.Client
)

type SignResult struct {
	Code string `json:"code"`
	Data struct {
		SignedRan    string `json:"signedRan"`
		Status       string `json:"status"`
		BeanUserType int    `json:"beanUserType"`
		AwardType    string `json:"awardType"`
		DailyAward   struct {
			Type      string `json:"type"`
			Title     string `json:"title"`
			SubTitle  string `json:"subTitle"`
			BeanAward struct {
				BeanCount  string `json:"beanCount"`
				BeanImgURL string `json:"beanImgUrl"`
			} `json:"beanAward"`
		} `json:"dailyAward"`
		ContinuityAward struct {
			Type      string `json:"type"`
			Title     string `json:"title"`
			SubTitle  string `json:"subTitle"`
			BeanAward struct {
				BeanCount  string `json:"beanCount"`
				BeanImgURL string `json:"beanImgUrl"`
			} `json:"beanAward"`
		} `json:"continuityAward"`
		ConductionBtn struct {
			BtnText string `json:"btnText"`
			LinkURL string `json:"linkUrl"`
		} `json:"conductionBtn"`
		SignRemind struct {
			Title        string `json:"title"`
			Content      string `json:"content"`
			PopImgURL    string `json:"popImgUrl"`
			BeanHomeLink string `json:"beanHomeLink"`
		} `json:"signRemind"`
		SignCalendar struct {
			CurrentDate    string `json:"currentDate"`
			SignRecordList []struct {
				Day         string `json:"day"`
				AwardState  string `json:"awardState"`
				AwardType   string `json:"awardType,omitempty"`
				AwardImgURL string `json:"awardImgUrl,omitempty"`
				AwardDesc   string `json:"awardDesc,omitempty"`
				PopText     string `json:"popText,omitempty"`
			} `json:"signRecordList"`
		} `json:"signCalendar"`
		Recommend struct {
		} `json:"recommend"`
		MsgGuideSwitch    string `json:"msgGuideSwitch"`
		SourceTips        string `json:"sourceTips"`
		TomorrowSendBeans int    `json:"tomorrowSendBeans"`
	} `json:"data"`
}

func init() {
	typefac.RegisterType(Sign{})
	log.Println("京东签到合集")
}

func (c Sign) Run() {
	var data = Redis.Keys(ctx, "baipiao:ck:*")
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
}

// 签到领京豆
func beanSign(c *resty.Request, user string) {
	result := new(SignResult)
	url := "https://api.m.jd.com/client.action?functionId=signBeanIndex&appid=ld"
	resp, err := c.SetResult(&result).Post(url)
	log.Println(string(resp.Body()), err)
}

func jdShop(c *resty.Request, name, data, user string) {
	url := "https://api.m.jd.com/?client=wh5&functionId=qryH5BabelFloors"
	resp, _ := c.SetHeader("Content-type", "application/x-www-form-urlencoded").SetFormData(map[string]string{
		"body": data,
	}).Post(url)
	body := string(resp.Body())
	floatLayerText := json.Get(body, "floatLayerList.#.params").Array()
	for _, result := range floatLayerText {
		if json.Get(result.String(), "enActK").String() != "" {
			//params:=
		}
	}
	floorList := json.Get(body, "floorList").Array()
	for _, result := range floorList {
		if result.Map()["template"].String() == "signIn" {
			signInfo := result.Map()["signInfos"]
			if signInfo.Map()["signStat"].String() == "1" {
				log.Printf("%s,%s今日已签到!", name, user)
			} else {
				log.Println(signInfo.Map()["params"])
			}
		}
	}
}

func jdShopSign(c *resty.Request, user string, name, tid string) {

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
