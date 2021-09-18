package task

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	. "github.com/teeoo/baipiao/cache"
	. "github.com/teeoo/baipiao/cron"
	"github.com/teeoo/baipiao/typefac"
	"golang.org/x/crypto/bcrypt"
	"log"
	"strconv"
	"time"
	"unsafe"
)

var (
	ctx = context.Background()
)

type School interface {
	Run()
}

func init() {
	s := Redis.Keys(ctx, "baipiao:auth")
	if len(s.Val()) == 0 {
		log.Println("正在为你初始化用户名和密码")
		password := Encode(strconv.FormatInt(time.Now().Unix(), 10))
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
		h := md5.New()
		h.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
		username := hex.EncodeToString(h.Sum(nil))[:5]
		Redis.Set(ctx, "baipiao:auth", fmt.Sprintf("%s:%s", username, string(hash)), 0)
		log.Printf("用户名为:%s 密码为:%s\n", username, password)
	}
	// 加载任务
	var data = Redis.Keys(ctx, "baipiao:cron:*")
	for _, v := range data.Val() {
		val := Redis.HGetAll(ctx, v)
		var _, spec, jobName = val.Val()["id"], val.Val()["spec"], val.Val()["jobName"]
		job, err := Task.AddJob(spec, typefac.CreateInstance(jobName, nil).(School))
		if err != nil {
			log.Println(err)
		}
		Redis.HSet(ctx, v, "id", strconv.Itoa(int(job)))
	}
}

func Encode(data string) string {
	content := *(*[]byte)(unsafe.Pointer(&data))
	coder := base64.NewEncoding("IJjkKLMNO567PQX12RVW3YZaDEFGbcdefghiABCHlSTUmnopqrxyz04stuvw89+/")
	return coder.EncodeToString(content)
}
