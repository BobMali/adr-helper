package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/BobMali/adr-helper/internal/adr"
	"github.com/BobMali/adr-helper/internal/web"
	webui "github.com/BobMali/adr-helper/web"
)

// configScopeStore persists scope-vocabulary additions to .adr.json under a
// mutex, keeping the in-memory config consistent for concurrent requests.
type configScopeStore struct {
	mu  sync.Mutex
	dir string
	cfg *adr.Config
}

func (s *configScopeStore) Scopes() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.cfg.Scopes...)
}

func (s *configScopeStore) AddScope(value string) ([]string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	prev := s.cfg.Scopes
	scopes, err := s.cfg.AddScope(value)
	if err != nil {
		return nil, err
	}
	if err := adr.SaveConfig(s.dir, s.cfg); err != nil {
		s.cfg.Scopes = prev // roll back the in-memory mutation to match disk
		return nil, err
	}
	return scopes, nil
}

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
		opts = append(opts, web.WithSuperseder(fileRepo))
		opts = append(opts, web.WithRelator(fileRepo))
		opts = append(opts, web.WithContentUpdater(fileRepo))

		// Auto-discover scopes from existing ADRs into the served vocabulary.
		// In-memory only: no config write at boot (safe on read-only mounts and
		// multi-replica deploys). Persistence stays with `adr init` / `adr scope
		// discover`, and lazily via the next web AddScope which saves the config.
		added, invalid, derr := adr.DiscoverAndMergeScopes(cfg)
		if derr != nil {
			log.Printf("warning: scope discovery failed: %v", derr)
		}
		for _, v := range invalid {
			log.Printf("warning: skipped invalid scope %q from ADRs", v)
		}
		if len(added) > 0 {
			log.Printf("discovered %d scope(s) from existing ADRs (in-memory; run 'adr scope discover' to persist): %s",
				len(added), strings.Join(added, ", "))
		}

		opts = append(opts, web.WithConfig(cfg))
		opts = append(opts, web.WithScopeStore(&configScopeStore{dir: ".", cfg: cfg}))
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
