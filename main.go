package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/rovarghe/mule/internal"
	"github.com/rovarghe/mule/internal/builtin"
	"github.com/rovarghe/mule/schema"
)

type H struct {
	context context.Context
}

func (h *H) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	internal.Process(h.context, r)

	/*
		switch r.URL.Path {
		case "/":
			fmt.Println("RtmpOot served")
		default:
			http.DefaultServeMux.ServeHTTP(w, r)
		}
	*/
}

func startServer(ctx context.Context) error {
	fmt.Println("Listening on port", 8000)
	return http.ListenAndServe(":8000", &H{
		context: ctx,
	})
}

func main() {
	var modules = []schema.Module{
		builtin.CoreModule,
	}

	ctx, err := internal.LoadModules(context.Background(), modules)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	log.Fatal(startServer(ctx))
}
