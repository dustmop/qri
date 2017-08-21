package server

import (
	"fmt"
	"github.com/datatogether/api/apiutil"
	ipfs "github.com/qri-io/castore/ipfs"
	"github.com/qri-io/qri/core/datasets"
	"github.com/qri-io/qri/core/queries"
	"github.com/qri-io/qri/p2p"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
)

type Server struct {
	cfg *Config
	log *logrus.Logger

	qriNode *p2p.QriNode
	store   *ipfs.Datastore
}

func New(options ...func(*Config)) (*Server, error) {
	cfg := DefaultConfig()
	for _, opt := range options {
		opt(cfg)
	}
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("server configuration error: %s", err.Error())
	}

	s := &Server{
		cfg: cfg,
		log: logrus.New(),
	}

	// output to stdout in dev mode
	if s.cfg.Mode == DEVELOP_MODE {
		s.log.Out = os.Stdout
	} else {
		s.log.Out = os.Stderr
	}
	s.log.Level = logrus.InfoLevel
	s.log.Formatter = &logrus.TextFormatter{
		ForceColors: true,
	}

	return s, nil
}

// main app entry point
func (s *Server) Serve() error {
	if s.cfg.LocalIpfs {
		store, err := ipfs.NewDatastore(func(cfg *ipfs.StoreCfg) {
			cfg.Online = false
		})
		if err != nil {
			return err
		}
		s.store = store
	}

	qriNode, err := p2p.NewQriNode(func(ncfg *p2p.NodeCfg) {
		ncfg.RepoPath = s.cfg.QriRepoPath
	})
	if err != nil {
		return err
	}

	s.qriNode = qriNode

	fmt.Println("qri addresses:")
	for _, a := range s.qriNode.EncapsulatedAddresses() {
		fmt.Printf("  %s\n", a.String())
	}

	server := &http.Server{}
	server.Handler = s.NewServerRoutes()
	// fire it up!
	s.log.Println("starting server on port", s.cfg.Port)
	// http.ListenAndServe will not return unless there's an error
	return StartServer(s.cfg, server)
}

// NewServerRoutes returns a Muxer that has all API routes.
// This makes for easy testing using httptest, see server_test.go
func (s *Server) NewServerRoutes() *http.ServeMux {
	m := http.NewServeMux()

	m.HandleFunc("/", WebappHandler)
	m.Handle("/status", s.middleware(apiutil.HealthCheckHandler))

	m.Handle("/ipfs/", s.middleware(s.HandleIPFSPath))

	dsh := datasets.NewHandlers(s.store, s.qriNode.Repo())
	m.Handle("/datasets", s.middleware(dsh.DatasetsHandler))
	m.Handle("/datasets/", s.middleware(dsh.DatasetHandler))
	m.Handle("/data/ipfs/", s.middleware(dsh.StructuredDataHandler))

	qh := queries.NewHandlers(s.store, s.qriNode.Repo())
	m.Handle("/run", s.middleware(qh.RunHandler))

	return m
}