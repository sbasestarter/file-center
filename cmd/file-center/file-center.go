package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sbasestarter/file-center/internal/file-center/server"
	filepb "github.com/sbasestarter/proto-repo/gen/protorepo-file-go"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libconfig"
	"github.com/sgostarter/liblogrus"
	"github.com/sgostarter/librediscovery"
	"github.com/sgostarter/libservicetoolset/dbtoolset"
	"github.com/sgostarter/libservicetoolset/servicetoolset"
	"google.golang.org/grpc"
)

func main() {
	logger := l.NewWrapper(liblogrus.NewLogrus())
	logger.GetLogger().SetLevel(l.LevelDebug)

	var cfg config.Config

	_, err := libconfig.Load("file-center.yml", &cfg)
	if err != nil {
		logger.Fatalf("load config failed: %v", err)

		return
	}

	cfg.Logger = logger
	cfg.ContextLogger = logger.GetWrapperWithContext()

	if cfg.StgRoot == "" {
		cfg.StgRoot = "./"
	}

	cfg.StgRoot, _ = filepath.Abs(cfg.StgRoot)
	if cfg.StgTmpRoot == "" {
		cfg.StgTmpRoot = "./"
	}

	cfg.StgTmpRoot, _ = filepath.Abs(cfg.StgTmpRoot)

	dbToolset := dbtoolset.NewToolset(&cfg.DbConfig, logger)

	cfg.GRpcServerConfig.DiscoveryExConfig.Setter, err = librediscovery.NewSetter(context.Background(), logger,
		dbToolset.GetRedis(), "", time.Minute)
	if err != nil {
		logger.Fatalf("create rediscovery setter failed: %v", err)

		return
	}

	fileCenterServer := server.NewServer(context.Background(), &cfg, dbToolset)

	serviceToolset := servicetoolset.NewServerToolset(context.Background(), logger)

	_ = serviceToolset.CreateGRpcServer(&cfg.GRpcServerConfig, nil, func(s *grpc.Server) error {
		filepb.RegisterFileServiceServer(s, fileCenterServer)

		return nil
	})

	r := mux.NewRouter()
	fileCenterServer.HTTPRegister(r)
	cfg.HTTPServerConfig.Handler = r
	_ = serviceToolset.CreateHTTPServer(&cfg.HTTPServerConfig)

	serviceToolset.Wait()
}
