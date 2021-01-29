package config

import (
	"github.com/jiuzhou-zhao/go-fundamental/dbtoolset"
	"github.com/jiuzhou-zhao/go-fundamental/servicetoolset"
)

type Config struct {
	GRpcServerConfig servicetoolset.GRpcServerConfig
	HttpServerConfig servicetoolset.HttpServerConfig
	DbConfig         dbtoolset.DBConfig

	StgRoot    string
	StgTmpRoot string
}
