package jd

import (
	"github.com/teeoo/baipiao/typefac"
	"log"
)

type Bean struct{}

func init() {
	typefac.RegisterType(Bean{})
	log.Println("京东APP->京东到家->签到->所有任务")
}

func (c Bean) Run() {
	log.Println("JdBean")
}
