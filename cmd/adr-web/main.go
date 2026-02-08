package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/malek/adr-helper/internal/web"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	srv := web.NewServer()

	fmt.Fprintf(os.Stdout, "adr-web listening on %s\n", *addr)
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatal(err)
	}
}
