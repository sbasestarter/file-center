package main

import (
	"context"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/sbasestarter/file-center/internal/config"
	"github.com/sbasestarter/file-center/internal/file-center/server"
	"github.com/sbasestarter/proto-repo/gen/protorepo-file-center-go"
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libconfig"
	"github.com/sgostarter/libeasygo/stg"
	"github.com/sgostarter/liblogrus"
	"github.com/sgostarter/librediscovery"
	"github.com/sgostarter/libservicetoolset/servicetoolset"
	"google.golang.org/grpc"
)

func main() {
	logger := l.NewWrapper(liblogrus.NewLogrus())
	logger.GetLogger().SetLevel(l.LevelDebug)

	var cfg config.Config
	_, err := libconfig.Load("file_svr.yml", &cfg)
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

	redisCli, err := stg.InitRedis(cfg.RedisDSN)
	if err != nil {
		panic(err)
	}

	cfg.GRpcServerConfig.DiscoveryExConfig.Setter, err = librediscovery.NewSetter(context.Background(), logger, redisCli,
		"", time.Minute)
	if err != nil {
		logger.Fatalf("create rediscovery setter failed: %v", err)
		return
	}

	fileCenterServer := server.NewServer(context.Background(), &cfg)
	serviceToolset := servicetoolset.NewServerToolset(context.Background(), logger)
	_ = serviceToolset.CreateGRpcServer(&cfg.GRpcServerConfig, nil, func(s *grpc.Server) error {
		filecenterpb.RegisterFileCenterServer(s, fileCenterServer)

		return nil
	})

	r := mux.NewRouter()
	fileCenterServer.HTTPRegister(r)
	cfg.HttpServerConfig.Handler = r
	_ = serviceToolset.CreateHTTPServer(&cfg.HttpServerConfig)

	serviceToolset.Wait()
}
