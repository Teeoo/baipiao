package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/robfig/cron/v3"
	json "github.com/tidwall/gjson"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	. "github.com/teeoo/baipiao/cache"
	Config "github.com/teeoo/baipiao/config"
	. "github.com/teeoo/baipiao/cron"
	"github.com/teeoo/baipiao/graph/generated"
	"github.com/teeoo/baipiao/graph/model"
	"github.com/teeoo/baipiao/typefac"
	"golang.org/x/crypto/bcrypt"
)

type ResultRefreshToken struct {
	Code     string `json:"code"`
	TokenKey string `json:"tokenKey"`
	URL      string `json:"url"`
}

func (r *mutationResolver) AddJdCookies(ctx context.Context, cookie model.InputCookie) (*model.Cookies, error) {
	Redis.HMSet(ctx, fmt.Sprintf("baipiao:ck:%s", cookie.PtPin), "pt_pin", cookie.PtPin, "pt_key", cookie.PtKey, "ws_key", *cookie.WsKey, "qq", *cookie.Qq, "remark", *cookie.Remark)
	var data = model.Cookies{
		PtKey:  cookie.PtKey,
		PtPin:  cookie.PtPin,
		WsKey:  cookie.WsKey,
		Remark: cookie.Remark,
	}
	return &data, nil
}

func (r *mutationResolver) CronAddJob(ctx context.Context, spec *string, cmd *string) (*int, error) {
	school := typefac.CreateInstance(fmt.Sprintf("jd.%s", *cmd), nil).(School)
	val, e := Redis.HGet(ctx, fmt.Sprintf("baipiao:cron:jd.%s", *cmd), "id").Int()
	if e == nil {
		Task.Remove(cron.EntryID(val))
	}
	job, err := Task.AddJob(*spec, school)
	if err != nil {
		return nil, err
	}
	Redis.HMSet(ctx, fmt.Sprintf("baipiao:cron:jd.%s", *cmd), "id", strconv.Itoa(int(job)), "spec", *spec, "jobName", fmt.Sprintf("jd.%s", *cmd))
	Task.Start()
	return func(val int) *int { return &val }(int(job)), nil
}

func (r *mutationResolver) CronDelJob(ctx context.Context, jobID *int) (*int, error) {
	Task.Remove(cron.EntryID(*jobID))
	return nil, nil
}

func (r *queryResolver) Login(ctx context.Context, user *string, pwd *string) (string, error) {
	value, _ := Redis.Get(ctx, "baipiao:auth").Result()
	err := bcrypt.CompareHashAndPassword([]byte(strings.Split(value, ":")[1]), []byte(*pwd))
	if err != nil {
		return "", errors.New("用户名或密码不正确")
	}
	if strings.Split(value, ":")[0] == *user && err == nil {
		token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
			IssuedAt:  time.Now().Unix(),
			Id:        *user,
		}).SignedString([]byte(Config.Config.Jwt.JWT_SECRET))
		return token, err
	}
	return "", err
}

func (r *queryResolver) GetJdCookies(ctx context.Context) ([]*model.Cookies, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CheckCookies(ctx context.Context) ([]*model.CheckCookies, error) {
	var check []*model.CheckCookies
	apiUrl := "https://api.m.jd.com/client.action?functionId=newUserInfo&clientVersion=10.0.9&client=android&openudid=a27b83d3d1dba1cc&uuid=a27b83d3d1dba1cc&aid=a27b83d3d1dba1cc&area=19_1601_36953_50397&st=1626848394828&sign=447ffd52c08f0c8cca47ebce71579283&sv=101&body=%7B%22flag%22%3A%22nickname%22%2C%22fromSource%22%3A1%2C%22sourceLevel%22%3A1%7D&"
	var data = Redis.Keys(ctx, "baipiao:ck:*")
	for _, s := range data.Val() {
		result := Redis.HGetAll(ctx, s)
		message := "有效"

		resp, _ := resty.New().R().SetCookies([]*http.Cookie{
			{
				Name:  "pt_pin",
				Value: result.Val()["pt_pin"],
			}, {
				Name:  "pt_key",
				Value: result.Val()["pt_key"],
			},
		}).
			SetHeader("User-Agent", "jdapp;iPhone;10.1.2;15.0;cc4a3fee7254710140e7ccc0443480e5d6b3ca68;network/wifi;model/iPhone12,1;addressid/2865568211;appBuild/167802;jdSupportDarkMode/0;Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148;supportJDSHWK/1").
			Post(apiUrl)
		name, _ := url.QueryUnescape(result.Val()["pt_pin"])
		client := resty.New()
		if json.Get(string(resp.Body()), "code").String() != "0" {
			message = "无效[已刷新]"
			// TODO:通过ws_key刷新pt_key
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
					"area":          "24_2144_24463_58345",
					"body":          `{"to":"https:\/\/plogin.m.jd.com\/jd-mlogin\/static\/html\/appjmp_blank.html","action":"to"}`,
					"build":         `167814`,
					"client":        `apple`,
					"clientVersion": "10.1.4",
					"d_brand":       "apple",
					"d_model":       "iPhone12,1",
					"eid":           "eidI46588123b9s3WNmzB1PXT6mwp+pzjqIUO1jZtHnWKawlX2malzeRTmEYkFhIKa7uFmZOsAplfKhmwEhP1pVEtTxxuF1WEejsCBMzayY7eJO0k/F5",
					"isBackground":  "N",
					//"joycious":"81",
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
					Redis.HSet(ctx, s, "pt_key", cookie.Value)
				}
			}
		}
		if result.Val()["remark"] != "" {
			name = result.Val()["remark"]
		}
		check = append(check, &model.CheckCookies{
			User:  fmt.Sprintf("账号:%s", name),
			Check: message,
		})
	}
	return check, nil
}

func (r *subscriptionResolver) Log(ctx context.Context, jobID int) (<-chan string, error) {
	panic(fmt.Errorf("not implemented"))
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

// Subscription returns generated.SubscriptionResolver implementation.
func (r *Resolver) Subscription() generated.SubscriptionResolver { return &subscriptionResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
type subscriptionResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.

type newUserInfo struct {
	Code           string `json:"code"`
	Uts            string `json:"uts"`
	UserPlusStatus bool   `json:"userPlusStatus"`
	UserInfoSns    string `json:"userInfoSns"`
	Enc            int    `json:"enc"`
	NoModifyText   string `json:"noModifyText"`
	CloseReminder  struct {
		CardSubTitle  string `json:"cardSubTitle"`
		TempCardTitle string `json:"tempCardTitle"`
		CardTitle     string `json:"cardTitle"`
	} `json:"closeReminder"`
}
