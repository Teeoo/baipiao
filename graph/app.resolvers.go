package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"fmt"
	"github.com/guonaihong/gout"
	cron "github.com/robfig/cron/v3"
	"github.com/teeoo/baipiao/graph/generated"
	"github.com/teeoo/baipiao/graph/model"
	"github.com/teeoo/baipiao/typefac"
	"net/http"
	"net/url"
)

func (r *mutationResolver) CronAddJob(ctx context.Context, spec *string, cmd *string) (*int, error) {
	school := typefac.CreateInstance(fmt.Sprintf("jd.%s", *cmd), nil).(School)
	job, err := c.AddJob(*spec, school)
	if err != nil {
		return nil, err
	}
	c.Start()
	return func(val int) *int { return &val }(int(job)), nil
}

func (r *mutationResolver) CronDelJob(ctx context.Context, jobID *int) (*int, error) {
	c.Remove(cron.EntryID(*jobID))
	return nil, nil
}

func (r *queryResolver) GetJdCookies(ctx context.Context) ([]*model.Cookies, error) {
	panic(fmt.Errorf("not implemented"))
}

func (r *queryResolver) CheckCookies(ctx context.Context) ([]*model.CheckCookies, error) {
	var check []*model.CheckCookies
	apiUrl := "https://api.m.jd.com/client.action?functionId=newUserInfo&clientVersion=10.0.9&client=android&openudid=a27b83d3d1dba1cc&uuid=a27b83d3d1dba1cc&aid=a27b83d3d1dba1cc&area=19_1601_36953_50397&st=1626848394828&sign=447ffd52c08f0c8cca47ebce71579283&sv=101&body=%7B%22flag%22%3A%22nickname%22%2C%22fromSource%22%3A1%2C%22sourceLevel%22%3A1%7D&"
	for _, datum := range data {
		message := "有效"
		userInfo := new(newUserInfo)
		_ = gout.GET(apiUrl).
			SetCookies(&http.Cookie{
				Name:  "pt_key",
				Value: datum.PtKey,
			}, &http.Cookie{
				Name:  "pt_pin",
				Value: datum.PtPin,
			}).
			BindJSON(&userInfo).
			SetHeader(gout.H{
				"user-agen": "jdapp;iPhone;10.1.2;15.0;cc4a3fee7254710140e7ccc0443480e5d6b3ca68;network/wifi;model/iPhone12,1;addressid/2865568211;appBuild/167802;jdSupportDarkMode/0;Mozilla/5.0 (iPhone; CPU iPhone OS 15_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Mobile/15E148;supportJDSHWK/1",
			}).
			Do()
		name, _ := url.QueryUnescape(datum.PtPin)
		if userInfo.Code != "0" {
			message = "无效"
		}
		check = append(check, &model.CheckCookies{
			User:  fmt.Sprintf("账号:%s", name),
			Check: message,
		})
	}
	return check, nil
}

// Mutation returns generated.MutationResolver implementation.
func (r *Resolver) Mutation() generated.MutationResolver { return &mutationResolver{r} }

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type mutationResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }

// !!! WARNING !!!
// The code below was going to be deleted when updating resolvers. It has been copied here so you have
// one last chance to move it out of harms way if you want. There are two reasons this happens:
//  - When renaming or deleting a resolver the old code will be put in here. You can safely delete
//    it when you're done.
//  - You have helper methods in this file. Move them out to keep these resolver files clean.

var c = cron.New(cron.WithParser(cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)))
var data []*model.Cookies

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
