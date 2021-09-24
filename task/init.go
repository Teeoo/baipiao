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
	"io"
	"log"
	"os"
	"strconv"
	"time"
	"unsafe"
)

var (
	ctx    = context.Background()
	logger *log.Logger
)

type School interface {
	Run()
}

func init() {
	_, err := os.Stat("./logs/taskInit")
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll("./logs/taskInit", os.ModePerm)
		if err != nil {
			log.Printf("mkdir failed![%v]\n", err)
		}
	}
	file := fmt.Sprintf("%s/%s.log", "./logs/taskInit", time.Now().Format("2006-01-02-15-04-05"))
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	logger = log.New(io.MultiWriter(os.Stdout, logFile), "[init] ", log.Ldate|log.Ltime|log.Llongfile|log.Lshortfile)
	s := Redis.Keys(ctx, "baipiao:auth")
	if len(s.Val()) == 0 {
		logger.Println("正在为你初始化用户名和密码")
		password := Encode(strconv.FormatInt(time.Now().Unix(), 10))
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), 14)
		h := md5.New()
		h.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
		username := hex.EncodeToString(h.Sum(nil))[:5]
		Redis.Set(ctx, "baipiao:auth", fmt.Sprintf("%s:%s", username, string(hash)), 0)
		logger.Printf("用户名为:%s 密码为:%s\n", username, password)
	}
	// 加载原有任务
	var data = Redis.Keys(ctx, "baipiao:cron:*")
	for _, v := range data.Val() {
		val := Redis.HGetAll(ctx, v)
		var _, spec, jobName = val.Val()["id"], val.Val()["spec"], val.Val()["jobName"]
		logger.Printf("重启加载任务: %s 执行时间 %s\n", jobName, spec)
		job, err := Task.AddJob(spec, typefac.CreateInstance(jobName, nil).(School))
		if err != nil {
			logger.Println(err)
		}
		Task.Start()
		Redis.HSet(ctx, v, "id", strconv.Itoa(int(job)))
	}
	// ast.ScanFuncDeclByComment("/Users/lee/go/src/github.com/teeoo/baipiao/jd/jd_sign.go","","@Cron")
	if len(data.Val()) == 0 {
		//TODO:初始化时加载任务
	}
}

func Encode(data string) string {
	content := *(*[]byte)(unsafe.Pointer(&data))
	coder := base64.NewEncoding("IJjkKLMNO567PQX12RVW3YZaDEFGbcdefghiABCHlSTUmnopqrxyz04stuvw89+/")
	return coder.EncodeToString(content)
}
