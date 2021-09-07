package main

import (
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/teeoo/baipiao/graph"
	"github.com/teeoo/baipiao/graph/generated"
	_ "github.com/teeoo/baipiao/jd"
	"log"
	"net/http"
)

func main() {
	http.Handle("/graphql", handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}})))
	log.Fatal(http.ListenAndServe(":1234", nil))
}
