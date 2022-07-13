package config

import (
	"github.com/sgostarter/i/l"
	"github.com/sgostarter/libservicetoolset/dbtoolset"
	"github.com/sgostarter/libservicetoolset/servicetoolset"
)

type Config struct {
	GRpcServerConfig servicetoolset.GRPCServerConfig `yaml:"grpc_server_config"`
	HTTPServerConfig servicetoolset.HTTPServerConfig `yaml:"http_server_config"`
	DbConfig         dbtoolset.Config                `yaml:"db_config"`

	StgRoot    string `yaml:"stg_root"`
	StgTmpRoot string `yaml:"stg_tmp_root"`

	Logger        l.Wrapper            `yaml:"-" ignored:"true"`
	ContextLogger l.WrapperWithContext `yaml:"-" ignored:"true"`
}
