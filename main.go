package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/teeoo/baipiao/gqlgen"
	"github.com/teeoo/baipiao/graph/generated"
	_ "github.com/teeoo/baipiao/jd"
	. "github.com/teeoo/baipiao/middleware"
	"log"
	"net/http"
)

func main() {
	http.Handle("/graphql", Jwt(handler.NewDefaultServer(generated.NewExecutableSchema(gqlgen.New()))))
	log.Fatal(http.ListenAndServe(":1234", nil))
}
