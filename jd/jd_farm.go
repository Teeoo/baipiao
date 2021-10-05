package jd

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	. "github.com/teeoo/baipiao/cache"
	. "github.com/teeoo/baipiao/http"
	"github.com/teeoo/baipiao/typefac"
	json "github.com/tidwall/gjson"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

type Farm struct{}

var FarmShareCode []map[string]string

func init() {
	PathExists("./logs/jd_farm")
	typefac.RegisterType(Farm{})
	log.Println("京东APP-我的->东东农场")
}

// Run @Cron 15 6-18/6 * * *
func (c Farm) Run() {
	loggerFile, err := os.OpenFile(fmt.Sprintf("%s/%s.log", "./logs/jd_farm", time.Now().Format("2006-01-02-15-04-05")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	log.SetOutput(io.MultiWriter(os.Stdout, loggerFile))
	log.SetPrefix(fmt.Sprintf("[%s]", "东东农场"))
	log.SetFlags(log.Ldate | log.Ltime | log.Llongfile | log.Lshortfile)
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
			if initFarm(HttpClient.R(), result.Val()["pt_pin"]) != "" {
				farmDailyTask(HttpClient.R(), result.Val()["pt_pin"])
				farmTenWater(HttpClient.R(), result.Val()["pt_pin"])
				farmFirstWaterAward(HttpClient.R(), result.Val()["pt_pin"])
				farmTenWaterAward(HttpClient.R(), result.Val()["pt_pin"])
				clickDuck(HttpClient.R(), result.Val()["pt_pin"])
				gotWater(HttpClient.R(), result.Val()["pt_pin"])
			}
		}
		farmHelp(HttpClient.R())
	}()
}

func _farm(c *resty.Request, functionId, body string) string {
	params := url.Values{}
	params.Add("appid", "wh5")
	params.Add("functionId", functionId)
	params.Add("body", body)
	u := fmt.Sprintf("https://api.m.jd.com/client.action?%s", params.Encode())
	timer := time.NewTimer(1 * time.Second)
	select {
	case <-timer.C:
		resp, err := c.SetHeaders(map[string]string{
			"user-agent":       UserAgent(),
			"x-requested-with": "com.jingdong.app.mall",
			"sec-fetch-mode":   "cors",
			"origin":           "https://carry.m.jd.com",
			"sec-fetch-site":   "same-site",
			"referer":          "https://carry.m.jd.com/babelDiy/Zeus/3KSjXqQabiTuD1cJ28QskrpWoBKT/index.html",
		}).SetBody(body).Post(u)
		if err != nil {
			log.Println(err)
		}
		timer.Stop()
		return string(resp.Body())
	}
}

// 初始化农场数据
func initFarm(c *resty.Request, user string) string {
	resp := _farm(c, "initForFarm", "")
	if json.Get(resp, "code").String() != "0" {
		return ""
	}
	farmUserPro := json.Get(resp, "farmUserPro").String()
	FarmShareCode = append(FarmShareCode, map[string]string{"user": user, "code": json.Get(farmUserPro, "shareCode").String()})
	if json.Get(farmUserPro, "shareCode").String() == "" {
		return ""
	}
	log.Printf("%s 的互助码为:%s", user, json.Get(farmUserPro, "shareCode").String())
	return farmUserPro
}

// 签到
func farmSign(c *resty.Request, user string) {
	data := _farm(c, "signForFarm", "")
	if json.Get(data, "code").String() == "0" {
		log.Printf("%s, 签到成功, 已连续签到%s天!", user, json.Get(data, "signDay").String())
	} else if json.Get(data, "code").String() == "7" {
		log.Printf("%s, 今日已签到过!", user)
	} else {
		log.Printf("%s, 签到失败,%s", user, json.Get(data, "message").String())
	}
	if json.Get(data, "todayGotWaterGoalTask.canPop").Bool() {
		resp := _farm(c, "gotWaterGoalTaskForFarm", `{'type': 3}`)
		log.Println(resp)
		if json.Get(resp, "code").String() == "0" {
			log.Printf("%s, 被水滴砸中, 获得%sg水滴!", user, json.Get(resp, "addEnergy").String())
		}
	}
}

// 每日任务
func farmDailyTask(c *resty.Request, user string) {
	data := _farm(c, "taskInitForFarm", "")
	if json.Get(data, "code").String() != "0" {
		log.Printf("%s, 获取领水滴任务列表失败!", user)
		return
	}
	if !json.Get(data, "signInit.todaySigned").Bool() {
		farmSign(HttpClient.R(), user)
	} else {
		log.Printf("%s, 今日已签到, 已连续签到%s天!", user, json.Get(data, "signInit.totalSigned").String())
	}
	if !json.Get(data, "gotBrowseTaskAdInit.f").Bool() {
		tasks := json.Get(data, "gotBrowseTaskAdInit.userBrowseTaskAds").Array()
		//log.Println(tasks)
		func(t []json.Result) {
			for _, result := range t {
				taskName := json.Get(result.String(), "mainTitle")
				log.Printf("%s 正在进行浏览任务: 《%s》...", user, taskName.String())
				taskResp := _farm(c, "browseAdTaskForFarm", fmt.Sprintf(`{"advertId": %s, "type": 0}`, json.Get(result.String(), "advertId").String()))
				if json.Get(taskResp, "code").String() == "0" || json.Get(taskResp, "code").String() == "7" {
					taskAward := _farm(c, "browseAdTaskForFarm", fmt.Sprintf(`{"advertId": "%s", "type": 1}`, json.Get(result.String(), "advertId").String()))
					if json.Get(taskAward, "code").String() == "0" {
						log.Printf("%s,  成功领取任务:《%s》的奖励, 获得%sg水滴！", user, taskName, json.Get(taskAward, "amount").String())
					} else {
						log.Printf("%s, 领取任务:《%s》的奖励失败, %s", user, taskName, taskAward)
					}
				} else {
					log.Printf("%s, 浏览任务:《%s》结果, %s", user, taskName, taskResp)
				}
			}
		}(tasks)
	} else {
		log.Printf("%s, 今日浏览广告任务已完成!", user)
	}
	if !json.Get(data, "gotThreeMealInit.f").Bool() {
		farmTimedCollarDrop(c, user)
	}
	if !json.Get(data, "waterFriendTaskInit.f").Bool() && json.Get(data, "waterFriendTaskInit.waterFriendCountKey").Int() < json.Get(data, "waterFriendTaskInit.waterFriendMax").Int() {
		farmFriendWater(c, user)
	}

	farmClockIn(c, user)

	farmWaterDropRain(c, user, json.Get(data, "waterRainInit").String())

	getExtraAward(c, user)

	turntable(c, user)

	park(c, user)

}

func park(c *resty.Request, user string) {
	resp := _farm(c, "ddnc_farmpark_Init", `{"version":1,"channel":1}`)
	if json.Get(resp, "code").String() != "0" {
		log.Printf("%s 无法获取东东乐园任务", user)
		return
	}
	for k, result := range json.Get(resp, "buildings").Array() {
		task := json.Get(result.String(), "topResource.task")
		if json.Get(task.String(), "status").Int() != 1 {
			log.Printf("%s 今日已完成东东乐园:%s 浏览任务!", user, json.Get(result.String(), "name"))
		} else {
			data := _farm(c, "ddnc_farmpark_markBrowser", fmt.Sprintf(`{"version":"1","channel":1,"advertId":"%s"}`, json.Get(task.String(), "advertId")))
			if json.Get(data, "code").String() != "0" {
				log.Printf("%s 无法进行东东乐园:%s 浏览任务, 原因:%s", user, json.Get(result.String(), "name"), json.Get(data, "message"))
			}
			log.Printf("%s 正在进行东东乐园:%s 浏览任务 %s", user, json.Get(result.String(), "name"), json.Get(task.String(), "browseSeconds"))
			browseAward := _farm(c, "ddnc_farmpark_browseAward", fmt.Sprintf(`{"version":"1","channel":1,"advertId":"%s","type":1,"index":"%s"}`, json.Get(task.String(), "advertId"), strconv.Itoa(k)))
			if json.Get(browseAward, "code").String() == "0" {
				log.Printf("%s 领取东东乐园:%s 浏览任务奖励成功,获得%sg水滴!", user, json.Get(result.String(), "name"), json.Get(browseAward, "result.waterEnergy").String())
			} else {
				log.Printf("%s 领取东东乐园:%s 浏览任务奖励失败 %s", user, json.Get(result.String(), "name"), json.Get(browseAward, "message").String())
			}
		}
	}
}

func turntable(c *resty.Request, user string) {
	resp := _farm(c, "initForTurntableFarm", "")
	log.Println(resp)
	if json.Get(resp, "code").String() != "0" {
		log.Printf("%s, 当前无法参与天天抽奖", user)
		return
	}
	if !json.Get(resp, "timingGotStatus").Bool() {
		if json.Get(resp, "sysTime").Int() > (json.Get(resp, "timingLastSysTime").Int() + 60*60*json.Get(resp, "timingIntervalHours").Int()*1000) {
			data := _farm(c, "timingAwardForTurntableFarm", "")
			log.Printf("%s 领取定时奖励结果 %s", user, data)
		} else {
			log.Printf("%s 免费赠送的抽奖机会未到时间", user)
		}
	} else {
		log.Printf("%s 4小时候免费赠送的抽奖机会已领取", user)
	}

	if len(json.Get(resp, "turntableBrowserAds").Array()) > 0 {
		for _, result := range json.Get(resp, "turntableBrowserAds").Array() {
			if result.Map()["status"].Bool() {
				log.Printf("%s 天天抽奖任务:%s, 今日已完成过!", user, result.Map()["main"].String())
				continue
			}
			browser := _farm(c, "browserForTurntableFarm", fmt.Sprintf(`{"type":1,"adId":"%s"}`, result.Map()["adId"].String()))
			log.Printf("%s 完成天天抽奖任务:《%s》, 结果:%s", user, result.Map()["main"].String(), browser)
			awardRes := _farm(c, "browserForTurntableFarm", fmt.Sprintf(`{"type":2,"adId":"%s"}`, result.Map()["adId"].String()))
			log.Printf("%s 领取天天抽奖任务:《%s》奖励, 结果:%s", user, result.Map()["main"].String(), awardRes)
		}
	}
	data := _farm(c, "initForTurntableFarm", "")
	lotteryTimes := json.Get(data, "remainLotteryTimes").Int()
	if lotteryTimes == 0 {
		log.Printf("%s 天天抽奖次数已用完, 无法抽奖！", user)
		return
	}
	log.Printf("%s 天天抽奖次数已用完, 无法抽奖！", user)
	for i := range make([]int, lotteryTimes+1) {
		lottery := _farm(c, "lotteryForTurntableFarm", "")
		log.Printf("%s 第%s 抽奖结果 %s", user, strconv.Itoa(i), lottery)
	}
}

func getExtraAward(c *resty.Request, user string) {
	resp := _farm(c, "masterHelpTaskInitForFarm", "")
	if !json.Get(resp, "masterHelpPeoples").Bool() && len(json.Get(resp, "masterHelpPeoples").Array()) < 5 {
		log.Printf("%s, 获取助力信息失败或者助力不满5人, 无法领取额外奖励!", user)
		return
	}
	awardRes := _farm(c, "masterGotFinishedTaskForFarm", "")
	if json.Get(awardRes, "code").String() == "0" {
		log.Printf("%s 成功领取好友助力奖励, %sg水滴", user, json.Get(awardRes, "amount").String())
	} else {
		log.Printf("%s 领取好友助力奖励失败 %s", user, awardRes)
	}
}

func farmWaterDropRain(c *resty.Request, user, task string) {
	if json.Get(task, "f").Bool() {
		log.Printf("%s, 两次水滴雨任务已全部完成", user)
	}

	if time.Now().Unix()*1000 < json.Get(task, "lastTime").Int()+3*60*60*1000 {
		log.Printf("%s 第%s次水滴雨未到时间:%s", user, strconv.Itoa(int(json.Get(task, "winTimes").Int()+1)), task)
		return
	}
	maxLimit := json.Get(task, "config.maxLimit").Int()
	for range make([]int, maxLimit) {
		resp := _farm(c, "waterRainForFarm", "")
		if json.Get(resp, "code").String() == "0" {
			log.Printf("%s 第%s次水滴雨获得水滴:%sg", user, strconv.Itoa(int(json.Get(task, "winTimes").Int()+1)), json.Get(resp, "addEnergy"))
		} else {
			log.Printf("%s 第%s次水滴雨执行错误:%s", user, strconv.Itoa(int(json.Get(task, "winTimes").Int()+1)), resp)
		}
	}
}

func farmClockIn(c *resty.Request, user string) {
	log.Printf("%s 开始打卡领水活动(签到, 关注)", user)
	resp := _farm(c, "clockInInitForFarm", "")
	if json.Get(resp, "code").String() == "0" {
		if !json.Get(resp, "todaySigned").Bool() {
			log.Printf("%s 开始今日签到", user)
			data := _farm(c, "clockInForFarm", `{"type":1}`)
			log.Printf("%s 签到结果 %s", user, data)
			if json.Get(data, "signDay").Int() == 7 {
				log.Printf("%s 开始领取--惊喜礼包!", user)
				giftData := _farm(c, "clockInForFarm", `{"type":2}`)
				log.Printf("%s 惊喜礼包获得%sg水滴", user, json.Get(giftData, "amount"))
			}
		}
	}
	if json.Get(resp, "todaySigned").Bool() && json.Get(resp, "totalSigned").Int() == 7 {
		log.Printf("%s 开始领取--惊喜礼包!", user)
		giftData := _farm(c, "clockInForFarm", `{"type":2}`)
		if json.Get(giftData, "code").String() == "7" {
			log.Printf("%s 领取惊喜礼包失败, 已领取过!", user)
		} else if json.Get(giftData, "code").String() == "0" {
			log.Printf("%s 惊喜礼包获得%sg水滴", user, json.Get(giftData, "amount"))
		} else {
			log.Printf("%s 领取惊喜礼包失败 %s", user, giftData)
		}
	}
	if json.Get(resp, "themes").IsArray() && len(json.Get(resp, "themes").Array()) > 0 {
		for _, result := range json.Get(resp, "themes").Array() {
			if !json.Get(result.String(), "hadGot").Bool() {
				log.Printf("%s 关注ID %s", user, json.Get(result.String(), "id").String())
				follow := _farm(c, "clockInFollowForFarm", fmt.Sprintf(`{"id":"%s","type":"theme","step":1}`))
				if json.Get(follow, "code").String() == "0" {
					_follow := _farm(c, "clockInFollowForFarm", fmt.Sprintf(`{"id":"%s","type":"theme","step":2}`))
					if json.Get(_follow, "code").String() == "0" {
						log.Printf("关注%s, 获得水滴%sg", json.Get(result.String(), "id").String(), json.Get(_follow, "amount").String())
					}
				}
			}
		}
	}
	log.Printf("%s 结束打卡领水活动(签到, 关注)", user)
}

func farmTimedCollarDrop(c *resty.Request, user string) {
	resp := _farm(c, "gotThreeMealForFarm", "")
	if json.Get(resp, "code").String() == "0" {
		log.Printf("%s, 【定时领水滴】获得 %sg!", user, json.Get(resp, "amount"))
	} else {
		log.Printf("%s, 【定时领水滴】失败 %s", user, resp)
	}
}

func farmFriendWater(c *resty.Request, user string) {
	resp := _farm(c, "friendListInitForFarm", "")
	if !json.Get(resp, "friends").IsArray() {
		log.Printf("%s, 获取好友列表失败 %s", user, resp)
		return
	}
	friends := json.Get(resp, "friends").Array()
	if len(friends) == 0 {
		log.Printf("%s 暂无好友", user)
	}
	var count = 0
	for _, friend := range friends {
		if json.Get(friend.String(), "friendState").Int() == 1 {
			count += 1
			share := _farm(c, "waterFriendForFarm", fmt.Sprintf(`{"shareCode":"%s"}`, json.Get(friend.String(), "shareCode").String()))
			log.Printf("%s 为第%s个好友(%s)浇水, 结果：%s", user, strconv.Itoa(count), json.Get(friend.String(), "nickName").String(), share)
			if json.Get(share, "code").String() == "11" {
				log.Printf("%s 水滴不够, 退出浇水!", user)
				return
			}
		}
	}
}

func farmTenWater(c *resty.Request, user string) {
	task := _farm(c, "taskInitForFarm", "")
	taskLimitTimes := json.Get(task, "totalWaterTaskInit.totalWaterTaskLimit").Int()
	curTimes := json.Get(task, "totalWaterTaskInit.totalWaterTaskTimes").Int()
	if curTimes == taskLimitTimes {
		log.Printf("%s 今日已完成十次浇水!", user)
		return
	}
	fruitFinished := false
	for i := curTimes; i < taskLimitTimes; i++ {
		log.Printf("%s 开始第 %s 次浇水", user, strconv.Itoa(int(i+1)))
		resp := _farm(c, "waterGoodForFarm", "")
		if json.Get(resp, "code").String() != "0" {
			log.Printf("%s 浇水异常 %s", user, resp)
			break
		}
		log.Printf("%s 剩余水滴 %s ", user, json.Get(resp, "totalEnergy").String())
		fruitFinished = json.Get(resp, "finished").Bool()
		if fruitFinished {
			break
		}
		if json.Get(resp, "totalEnergy").Int() < 10 {
			log.Printf("%s 水滴不够10g, 退出浇水!", user)
			break
		}
		farmStageAward(c, user, resp)
	}
	if fruitFinished {
		log.Printf("%s 水果已可领取", user)
	}
}

func farmStageAward(c *resty.Request, user, water string) {
	if json.Get(water, "waterStatus").Int() == 0 && json.Get(water, "treeEnergy").Int() == 10 {
		awardRes := _farm(c, "gotStageAwardForFarm", `{"type":"1"}`)
		log.Printf("%s 领取浇水第一阶段奖励: %s", user, awardRes)
	} else if json.Get(water, "waterStatus").Int() == 1 {
		awardRes := _farm(c, "gotStageAwardForFarm", `{"type":"2"}`)
		log.Printf("%s 领取浇水第二阶段奖励: %s", user, awardRes)
	} else if json.Get(water, "waterStatus").Int() == 2 {
		awardRes := _farm(c, "gotStageAwardForFarm", `{"type":"3"}`)
		log.Printf("%s 领取浇水第三阶段奖励: %s", user, awardRes)
	}
}

func farmFirstWaterAward(c *resty.Request, user string) {
	taskData := _farm(c, "taskInitForFarm", "")
	if !json.Get(taskData, "firstWaterInit.f").Bool() && json.Get(taskData, "firstWaterInit.totalWaterTimes").Int() > 0 {
		resp := _farm(c, "firstWaterTaskForFarm", "")
		if json.Get(resp, "code").String() == "0" {
			log.Printf("%s 【首次浇水奖励】获得%sg水滴!", user, json.Get(resp, "amount"))
		} else {
			log.Printf("%s 【首次浇水奖励】领取失败 %s", user, resp)
		}
	} else {
		log.Printf("%s 首次浇水奖励已领取", user)
	}
}

func farmTenWaterAward(c *resty.Request, user string) {
	taskData := _farm(c, "taskInitForFarm", "")
	taskLimitTimes := json.Get(taskData, "totalWaterTaskInit.totalWaterTaskLimit").Int()
	curTimes := json.Get(taskData, "totalWaterTaskInit.totalWaterTaskTimes").Int()
	if !json.Get(taskData, "totalWaterTaskInit.f").Bool() && curTimes >= taskLimitTimes {
		resp := _farm(c, "totalWaterTaskForFarm", "")
		if json.Get(resp, "code").String() == "0" {
			log.Printf("%s 【十次浇水奖励】获得%sg水滴!", user, json.Get(resp, "totalWaterTaskEnergy").String())
		} else {
			log.Printf("%s 【十次浇水奖励】领取失败 %s", user, resp)
		}
	} else if curTimes < taskLimitTimes {
		log.Printf("%s【十次浇水】任务未完成, 今日浇水:%s", user, strconv.Itoa(int(curTimes)))
	} else {
		log.Printf("%s【十次浇水】奖励已领取!", user)
	}
}

func farmHelp(c *resty.Request) {
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
		var helpMaxCount = 3
		var curCount = 0
		for i := 0; i < len(FarmShareCode); i++ {
			if curCount >= helpMaxCount {
				log.Printf("%s 今日助力次数已用完", result.Val()["pt_pin"])
			}
			ticker := time.NewTimer(2 * time.Second)
			select {
			case <-ticker.C:
				if result.Val()["pt_pin"] != FarmShareCode[i]["user"] {
					log.Printf(`账号%s去助力[%s]%s`, result.Val()["pt_pin"], FarmShareCode[i]["user"], FarmShareCode[i]["code"])
					resp := _farm(c, "initForFarm", fmt.Sprintf(`{"imageUrl":"","nickName":"","shareCode":"%s","babelChannel":"3"}`, FarmShareCode[i]["code"]))
					if json.Get(resp, "helpResult.code").String() == "0" {
						log.Printf("%s 已成功给【%s】助力!", result.Val()["pt_pin"], FarmShareCode[i]["user"])
						curCount += 1
					} else if json.Get(resp, "helpResult.code").String() == "9" {
						log.Printf("%s 之前给【%s】助力过了!", result.Val()["pt_pin"], FarmShareCode[i]["user"])
					} else if json.Get(resp, "helpResult.code").String() == "8" {
						log.Printf("%s 今日助力次数已用完", result.Val()["pt_pin"])
						break
					} else if json.Get(resp, "helpResult.code").String() == "10" {
						log.Printf("%s 好友 %s 已满五人助力", result.Val()["pt_pin"], FarmShareCode[i]["user"])
					} else {
						log.Printf("%s 给 %s 助力失败", result.Val()["pt_pin"], FarmShareCode[i]["user"])
					}
				}
			}
			ticker.Stop()
		}
	}
}

func clickDuck(c *resty.Request, user string) {
	for range [10]int{} {
		resp := _farm(c, "getFullCollectionReward", `{"type":2,"version":14,"channel":1,"babelChannel":0}`)
		if json.Get(resp, "code").String() == "0" {
			log.Printf("%s %s", user, json.Get(resp, "title").String())
		} else {
			log.Printf("%s 点鸭子次数已达上限!", user)
			break
		}
	}
}

func gotWater(c *resty.Request, user string) {
	resp := _farm(c, "gotWaterGoalTaskForFarm", `{"type":3,"version":14,"channel":1,"babelChannel":0}`)
	if json.Get(resp, "code").String() == "0" {
		log.Printf("%s 领取水滴 %s", user, resp)
	} else {
		log.Printf("%s 领取水滴失败 %s", user, resp)
	}
}
