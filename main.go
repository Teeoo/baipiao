package main

import (
	_ "github.com/teeoo/baipiao/gqlgen"
	_ "github.com/teeoo/baipiao/jd"
	_ "github.com/teeoo/baipiao/task"
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe(":1234", nil))
}
