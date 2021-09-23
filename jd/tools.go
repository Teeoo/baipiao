package jd

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"log"
	"math/big"
	random "math/rand"
	"os"
	"strconv"
	"time"
)

func PathExists(path string) {
	_, err := os.Stat(path)
	if err != nil && os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			log.Printf("mkdir failed![%v]\n", err)
		}
	}
}

func fp() string {
	e := "0123456789"
	a := 13
	i := ""
	for a > 0 {
		result, _ := rand.Int(rand.Reader, big.NewInt(int64(len(e))))
		i += fmt.Sprintf("%s", result)
		a -= 1
	}
	i += fmt.Sprintf("%s", strconv.FormatInt(time.Now().Unix()*100, 10))
	return i[0:16]
}

func GetRandomString(num int, str ...string) string {
	s := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	if len(str) > 0 {
		s = str[0]
	}
	l := len(s)
	r := random.New(random.NewSource(getRandSeek()))
	var buf bytes.Buffer
	for i := 0; i < num; i++ {
		x := r.Intn(l)
		buf.WriteString(s[x : x+1])
	}
	return buf.String()
}

func getRandSeek() int64 {
	l.Lock()
	if randSeek >= 100000000 {
		randSeek = 1
	}
	randSeek++
	l.Unlock()
	return time.Now().UnixNano() + randSeek

}

func initLogger(path, prefix string) *log.Logger {
	PathExists(path)
	loggerFile, err := os.OpenFile(fmt.Sprintf("%s/%s.log", path, time.Now().Format("2006-01-02-15-04-05")), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println(err)
	}
	//defer func(loggerFile *os.File) {
	//	err := loggerFile.Close()
	//	if err != nil {
	//		log.Println(err)
	//	}
	//}(loggerFile)
	//io.MultiWriter([]io.Writer{loggerFile, os.Stdout}...)
	//log.SetOutput(io.MultiWriter(os.Stdout, loggerFile))
	//log.SetOutput(io.MultiWriter(os.Stdout, loggerFile))
	//log.SetPrefix(fmt.Sprintf("[%s] ", prefix))
	//log.SetFlags(log.Ldate | log.Ltime | log.Llongfile | log.Lshortfile)
	return log.New(io.MultiWriter(os.Stdout, loggerFile), fmt.Sprintf("[%s] ", prefix), log.Ldate|log.Ltime|log.Llongfile|log.Lshortfile)
}
