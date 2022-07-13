package config

import (
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libservicetoolset/servicetoolset"
)

type Config struct {
	GRpcServerConfig servicetoolset.GRPCServerConfig `yaml:"grpc_server_config"`
	HttpServerConfig servicetoolset.HTTPServerConfig `yaml:"http_server_config"`
	RedisDSN         string                          `yaml:"redis_dsn"`

	StgRoot    string `yaml:"stg_root"`
	StgTmpRoot string `yaml:"stg_tmp_root"`

	Logger        l.Wrapper            `yaml:"-"`
	ContextLogger l.WrapperWithContext `yaml:"-"`
}
