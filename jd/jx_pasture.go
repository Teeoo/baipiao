package jd

import (
	j "encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	. "github.com/teeoo/baipiao/cache"
	. "github.com/teeoo/baipiao/http"
	"github.com/teeoo/baipiao/typefac"
	json "github.com/tidwall/gjson"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Pasture struct{}

type HomePageInfo struct {
	Data struct {
		Activeid  string `json:"activeid"`
		Activekey string `json:"activekey"`
		Avatar1   string `json:"avatar1"`
		Avatar2   string `json:"avatar2"`
		Cockinfo  struct {
			Createtime  int           `json:"createtime"`
			GiftIn      []interface{} `json:"gift_in"`
			GiftOut     []interface{} `json:"gift_out"`
			Matinglimit int           `json:"matinglimit"`
			Matingvalue int           `json:"matingvalue"`
			Petid       string        `json:"petid"`
			Status      int           `json:"status"`
			Type        int           `json:"type"`
		} `json:"cockinfo"`
		Coins          int           `json:"coins"`
		Cow            Cow           `json:"cow"`
		Dressinfo      []interface{} `json:"dressinfo"`
		Eggcnt         int           `json:"eggcnt"`
		Events         []interface{} `json:"events"`
		FinishedtaskId string        `json:"finishedtaskId"`
		Firstactive    string        `json:"firstactive"`
		Hatchboxinfo   struct {
			Currnum int `json:"currnum"`
			Status  int `json:"status"`
		} `json:"hatchboxinfo"`
		Hatchinfo       []interface{} `json:"hatchinfo"`
		Hj              string        `json:"hj"`
		Isactivenewuser int           `json:"isactivenewuser"`
		Ischangeactive  int           `json:"ischangeactive"`
		MaintaskId      string        `json:"maintaskId"`
		Materialinfo    []struct {
			Totalvalue int `json:"totalvalue"`
			Type       int `json:"type"`
			Usedvalue  int `json:"usedvalue"`
			Value      int `json:"value"`
		} `json:"materialinfo"`
		Newuserexchange int           `json:"newuserexchange"`
		Nickname        string        `json:"nickname"`
		Petinfo         []Petinfo     `json:"petinfo"`
		Pushnotice      bool          `json:"pushnotice"`
		Servetime       int64         `json:"servetime"`
		Sharekey        string        `json:"sharekey"`
		Toast           []interface{} `json:"toast"`
		Totalexperience int           `json:"totalexperience"`
	} `json:"data"`
	Message string `json:"message"`
	Ret     int    `json:"ret"`
}

type Cow struct {
	Currstage     int `json:"currstage"`
	Lastgettime   int `json:"lastgettime"`
	Nextstagecoin int `json:"nextstagecoin"`
	Perlimit      int `json:"perlimit"`
	Speed         int `json:"speed"`
	Totalcoin     int `json:"totalcoin"`
}

type Petinfo struct {
	Adult              int         `json:"adult"`
	Born               int         `json:"born"`
	Bornvalue          int         `json:"bornvalue"`
	Cangetborn         int         `json:"cangetborn"`
	Createtime         int         `json:"createtime"`
	Currgainexperience int         `json:"currgainexperience"`
	Currposition       int         `json:"currposition"`
	Doing              interface{} `json:"doing"`
	Exchangetime       int         `json:"exchangetime"`
	Experience         int         `json:"experience"`
	Hashatchegg        int         `json:"hashatchegg"`
	Lastborn           int         `json:"lastborn"`
	Old                int         `json:"old"`
	Petid              string      `json:"petid"`
	Progress           string      `json:"progress"`
	Requestid          string      `json:"requestid"`
	Sceneid            int         `json:"sceneid"`
	Sellcoin           int         `json:"sellcoin"`
	Source             int         `json:"source"`
	Stage              int         `json:"stage"`
	Status             int         `json:"status"`
	Type               int         `json:"type"`
}

var ShareCode []map[string]string

func init() {
	typefac.RegisterType(Pasture{})
	log.Println("京喜APP->京喜牧场->定时收金币/割草/投喂小鸡")
}

// Run @Cron 40 */1 * * *
func (c Pasture) Run() {
	var data = Redis.Keys(ctx, "baipiao:ck:*")
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
		homeData(HttpClient.R(), result.Val()["pt_pin"])
		goldFromBull(HttpClient.R(), result.Val()["pt_pin"])
		sign(HttpClient.R(), result.Val()["pt_pin"])
		dailyFood(HttpClient.R(), result.Val()["pt_pin"])
		buyFood(HttpClient.R(), result.Val()["pt_pin"])
		feed(HttpClient.R(), result.Val()["pt_pin"])
		mowing(HttpClient.R(), result.Val()["pt_pin"], 20)
		sweepChickenLegs(HttpClient.R(), result.Val()["pt_pin"], 8)
		tasks(HttpClient.R(), result.Val()["pt_pin"], 10)
	}
	help(HttpClient.R())
}

func homeData(c *resty.Request, user string) {
	data := request(c, "jxmc/queryservice/GetHomePageInfo", fmt.Sprintf(`{"isgift": "1","isquerypicksite": "1","_stk":"activeid,activekey,channel,isgift,isquerypicksite,sceneid"}`), user)
	homePageInfo := new(HomePageInfo)
	err := j.Unmarshal([]byte(data), homePageInfo)
	if err != nil {
		log.Println("首页数据获取出错", err)
	}
	coins = homePageInfo.Data.Coins
	activeId = homePageInfo.Data.Activeid
	petInfoList = homePageInfo.Data.Petinfo
	share_code = homePageInfo.Data.Sharekey
	cowInfo = homePageInfo.Data.Cow
	egg_num = homePageInfo.Data.Eggcnt
	curTaskStep = homePageInfo.Data.FinishedtaskId
	foodNum = 0
	if len(homePageInfo.Data.Materialinfo) != 0 {
		foodNum = homePageInfo.Data.Materialinfo[0].Value
	}
	ShareCode = append(ShareCode, map[string]string{"user": user, "code": homePageInfo.Data.Sharekey})
	log.Printf("%s 的互助码为:%s", user, homePageInfo.Data.Sharekey)
}

// 收牛的金币
func goldFromBull(c *resty.Request, user string) {
	data := request(c, "jxmc/operservice/GetCoin", fmt.Sprintf(`{"_stk": "activeid,activekey,channel,jxmc_jstoken,phoneid,sceneid,timestamp,token","token": "%s"}`, getToken(strconv.Itoa(cowInfo.Lastgettime))), user)
	if json.Get(data, "ret").Int() == 0 {
		log.Printf("%s 成功收牛牛, 获得金币:%s", user, json.Get(data, "data.addcoin").String())
	} else {
		log.Printf("%s 收取牛牛失败:%s", user, json.Get(data, "message").String())
	}
}

// 每天领白菜
func dailyFood(c *resty.Request, user string) {
	data := request(c, "/jxmc/operservice/GetVisitBackCabbage", fmt.Sprintf(`{"_stk": "activeid,activekey,channel,jxmc_jstoken,phoneid,sceneid,timestamp"}`), user)
	if json.Get(data, "ret").Int() == 0 {
		log.Printf("%s 成功领取白菜:%s", user, data)
	} else {
		log.Printf("%s 领取白菜失败:%s", user, json.Get(data, "message").String())
	}
}

// 买白菜
func buyFood(c *resty.Request, user string) {
	log.Printf("%s 当前白菜 %s 棵 当前金币 %s", user, strconv.Itoa(foodNum), strconv.Itoa(coins))
food:
	for foodNum <= 1000 && coins >= 5000 {
		ticker := time.NewTimer(1 * time.Second)
		select {
		case <-ticker.C:
			data := request(c, "jxmc/operservice/Buy", fmt.Sprintf(`{"_stk": "activeid,activekey,channel,jxmc_jstoken,phoneid,sceneid,timestamp,type","type":"1"}`), user)
			if json.Get(data, "ret").Int() == 200 {
				coins -= 5000
				foodNum += 100
				log.Printf("%s 成功购买白菜:%s", user, data)
			} else {
				log.Printf("%s 购买白菜失败:%s", user, json.Get(data, "message").String())
				ticker.Stop()
				break food
			}
		}
		ticker.Stop()
	}
}

// 投喂小🐔
func feed(c *resty.Request, user string) {
	if foodNum < 10 {
		log.Printf("%s当前白菜不足10棵,无法喂小鸡", user)
		return
	}
food:
	for foodNum >= 10 {
		ticker := time.NewTimer(2 * time.Second)
		select {
		case <-ticker.C:
			data := request(c, "jxmc/operservice/Feed", fmt.Sprintf(`{"_stk": "activeid,activekey,channel,jxmc_jstoken,phoneid,sceneid,timestamp"}`), user)
			if json.Get(data, "ret").Int() == 0 {
				log.Printf("%s 成功投喂一次小鸡:%s", user, data)
				foodNum = int(json.Get(data, "data.newnum").Int())
			} else if json.Get(data, "ret").Int() == 2020 && json.Get(data, "data.maintaskId").String() == "pause" {
				result := request(c, "jxmc/operservice/GetSelfResult", fmt.Sprintf(`{"_stk": "channel,itemid,sceneid,type","petid":"%s","type":"11"}`, petInfoList[0].Petid), user)
				if json.Get(result, "ret").Int() == 0 {
					log.Printf("%s 成功收取一枚金蛋, 当前金蛋:%s", user, json.Get(result, "data.newnum"))
				}
			} else {
				log.Printf("%s 投喂失败:%s", user, data)
				if json.Get(data, "ret").Int() == 2005 || json.Get(data, "ret").Int() == 2004 {
					log.Println("小鸡吃太饱了,或者任务未解锁")
					ticker.Stop()
					break food
				}
			}
		}
	}
}

// 割草
func mowing(c *resty.Request, user string, max int) {
mo:
	for i := 1; i <= max; i++ {
		ticker := time.NewTimer(1 * time.Second)
		select {
		case <-ticker.C:
			data := request(c, "jxmc/operservice/Action", fmt.Sprintf(`{"_stk": "activeid,activekey,channel,jxmc_jstoken,phoneid,sceneid,timestamp,type","type":"2"}`), user)
			if json.Get(data, "ret").Int() != 0 {
				log.Printf("%s 第 %s 次割草失败 %s", user, strconv.Itoa(i), data)
				break mo
			}
			log.Printf("%s 第 %s 次割草成功,获得金币 %s", user, strconv.Itoa(i), json.Get(data, "data.addcoins").String())
			if json.Get(data, "data.surprise").Bool() {
				result := request(c, "jxmc/operservice/GetSelfResult", fmt.Sprintf(`{"_stk": "activeid,activekey,channel,sceneid,type","type":"14"}`), user)
				if json.Get(result, "ret").Int() == 0 {
					log.Printf("%s 获得割草奖励 %s", user, json.Get(result, "data.prizepool").String())
				}
			}
		}
		ticker.Stop()
	}
}

// 签到
func sign(c *resty.Request, user string) {
	data := request(c, "jxmc/queryservice/GetSignInfo", fmt.Sprintf(`{"_stk"": "activeid,activekey,channel,sceneid"}`), user)
	if json.Get(data, "ret").Int() == 0 {
		for _, result := range json.Get(data, "data.signlist").Array() {
			if result.Map()["hasdone"].Bool() {
				log.Printf("%s 今日已签到", user)
			}
			res := request(c, "jxmc/operservice/GetSignReward", fmt.Sprintf(`{"_stk"": "channel,currdate,sceneid","currdate":"%s"}`, json.Get(data, "data.currdate")), user)
			if json.Get(res, "ret").Int() == 0 {
				log.Printf("%s 签到成功", user)
			} else {
				log.Printf("%s 签到失败:%s", user, json.Get(res, "message").String())
			}
		}
	} else {
		log.Printf("%s 获取签到数据失败:%s", user, json.Get(data, "message").String())
	}
}

// 任务
func tasks(c *resty.Request, user string, max int) {
	for i := 1; i <= max; i++ {
		//var flag = false
		result := request(c, "/newtasksys/newtasksys_front/GetUserTaskStatusList", fmt.Sprintf(`{"_stk": "bizCode,dateType,jxpp_wxapp_type,showAreaTaskFlag,source","source":"jxmc","bizCode":"jxmc","dateType":"","showAreaTaskFlag":"0","jxpp_wxapp_type":"7","gty":"ajax"}`), user)
		if json.Get(result, "ret").Int() != 0 {
			log.Printf("%s 获取每日任务列表失败 %s", user, result)
		}
		item := json.Get(result, "data.userTaskStatusList").Array()
		log.Println(item)
		for _, r := range item {
			ticker := time.NewTimer(1 * time.Second)
			select {
			case <-ticker.C:
				taskType, taskName := r.Map()["taskType"].Int(), r.Map()["taskName"].String()
				if r.Map()["awardStatus"].Int() == 1 {
					log.Printf("%s 奖励已领取 %s", user, taskName)
				}
				if r.Map()["completedTimes"].Int() >= r.Map()["targetTimes"].Int() {
					data := request(c, "/newtasksys/newtasksys_front/Award", fmt.Sprintf(`{"_stk": "bizCode,source,taskId","source":"jxmc","bizCode":"jxmc","gty":"ajax","taskId":"%s"}`, r.Map()["taskId"].String()), user)
					if json.Get(data, "ret").Int() == 0 {
						log.Printf("%s 成功领取任务《%s》奖励!", user, taskName)
					}
					time.Sleep(2 * time.Second)
				}
				if taskType == 2 {
					data := request(c, "/newtasksys/newtasksys_front/DoTask", fmt.Sprintf(`{"_stk": "bizCode,configExtra,source,taskId","source":"jxmc","bizCode":"jxmc","gty":"ajax","taskId":"%s","configExtra":""}`, r.Map()["taskId"].String()), user)
					if json.Get(data, "ret").Int() == 0 {
						log.Printf("%s 成功完成任务《%s》!", user, taskName)
					}
				}
			}
			ticker.Stop()
		}
	}
}

// 扫鸡腿
func sweepChickenLegs(c *resty.Request, user string, max int) {
chicken:
	for i := 1; i <= max; i++ {
		ticker := time.NewTimer(2 * time.Second)
		select {
		case <-ticker.C:
			data := request(c, "jxmc/operservice/Action", fmt.Sprintf(`{"_stk": "activeid,activekey,channel,petid,sceneid,type","type":"1","petid":"%s"}`, petInfoList[0].Petid), user)
			if json.Get(data, "ret").Int() != 0 {
				log.Printf("%s 第 %s 次扫鸡腿失败 %s", user, strconv.Itoa(i), data)
				break chicken
			}
			log.Printf("%s 第 %s 次扫鸡腿成功, 获得金币: %s", user, strconv.Itoa(i), json.Get(data, "data.addcoins").String())
			if json.Get(data, "data.surprise").Bool() {
				result := request(c, "jxmc/operservice/GetSelfResult", fmt.Sprintf(`{"_stk": "activeid,activekey,channel,sceneid,type","type":"14"}`), user)
				if json.Get(result, "ret").Int() == 0 {
					log.Printf("%s 获得割草奖励 %s", user, json.Get(result, "data.prizepool").String())
				}
			}
		}
		ticker.Stop()
	}
}

// 助力
func help(c *resty.Request) {
	var data = Redis.Keys(ctx, "baipiao:ck:*")
	for _, s := range data.Val() {
		result := Redis.HGetAll(ctx, s)
		c.SetCookies([]*http.Cookie{
			{
				Name:  "pt_pin",
				Value: result.Val()["pt_pin"],
			}, {
				Name:  "pt_key",
				Value: result.Val()["pt_key"],
			},
		})
		for i := 0; i < len(ShareCode); i++ {
			ticker := time.NewTimer(2 * time.Second)
			select {
			case <-ticker.C:
				if result.Val()["pt_pin"] != ShareCode[i]["user"] {
					log.Printf(`账号%s去助力%s`, result.Val()["pt_pin"], ShareCode[i]["code"])
					resp := request(c, "/jxmc/operservice/EnrollFriend", fmt.Sprintf(`{"_stk": "channel,sceneid,sharekey","sharekey":"%s"}`, ShareCode[i]["code"]), result.Val()["pt_pin"])
					log.Println(resp)
				}
			}
			ticker.Stop()
		}
	}
}
