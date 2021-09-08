package main

import (
	_ "github.com/teeoo/baipiao/gqlgen"
	_ "github.com/teeoo/baipiao/jd"
	"log"
	"net/http"
)

func main() {
	log.Fatal(http.ListenAndServe(":1234", nil))
}
