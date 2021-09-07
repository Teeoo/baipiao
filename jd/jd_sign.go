package jd

import (
	"github.com/teeoo/baipiao/typefac"
	"log"
)

type Sign struct{}

func init() {
	typefac.RegisterType(Sign{})
	log.Println("京东签到合集")
}

func (c Sign) Run() {
	log.Println("JdSignJob")
}
