package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/BobMali/adr-helper/internal/web"
	webui "github.com/BobMali/adr-helper/web"
)

func main() {
	addr := flag.String("addr", ":8080", "HTTP listen address")
	flag.Parse()

	var repo adr.Repository
	var opts []web.ServerOption
	cfg, err := adr.LoadConfig(".")
	if err != nil {
		log.Printf("warning: could not load config: %v (API will return 503)", err)
	} else {
		fileRepo := adr.NewFileRepository(cfg.Directory)
		repo = fileRepo
		opts = append(opts, web.WithStatusUpdater(fileRepo))
	}
	if subFS, err := fs.Sub(webui.DistFS, "dist"); err == nil {
		if _, err := subFS.Open("index.html"); err == nil {
			opts = append(opts, web.WithFrontend(subFS))
			log.Println("serving embedded frontend")
		}
	}

	srv := web.NewServer(repo, opts...)

	fmt.Fprintf(os.Stdout, "adr-web listening on %s\n", *addr)
	if err := srv.ListenAndServe(*addr); err != nil {
		log.Fatal(err)
	}
}
