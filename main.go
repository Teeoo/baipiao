package main

import (
	"fmt"
	. "github.com/teeoo/baipiao/config"
	_ "github.com/teeoo/baipiao/gqlgen"
	_ "github.com/teeoo/baipiao/jd"
	_ "github.com/teeoo/baipiao/task"
	"net/http"

	"log"
	"strconv"
)

func main() {
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", strconv.Itoa(Config.Web.Port)), nil))
}
