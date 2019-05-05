package main

import (
	"github.com/openfaas/faas-federation/handlers"
	"github.com/openfaas/faas-federation/routing"
	"github.com/openfaas/faas-federation/types"
	"github.com/openfaas/faas-federation/version"
	"github.com/openfaas/faas-provider"
	"github.com/openfaas/faas-provider/proxy"
	"os"
	"strings"

	bootTypes "github.com/openfaas/faas-provider/types"
	log "github.com/sirupsen/logrus"
)

func init() {
	logFormat := os.Getenv("LOG_FORMAT")
	logLevel := os.Getenv("LOG_LEVEL")
	if strings.EqualFold(logFormat, "json") {
		log.SetFormatter(&log.JSONFormatter{
			FieldMap: log.FieldMap{
				log.FieldKeyMsg:  "message",
				log.FieldKeyTime: "@timestamp",
			},
			TimestampFormat: "2006-01-02T15:04:05.999Z07:00",
		})
	} else {
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
	}

	if level, err := log.ParseLevel(logLevel); err == nil {
		log.SetLevel(level)
	}
}

func main() {

	log.Infof("faas-federation version:%s. Last commit message: %s, commit SHA: %s'", version.BuildVersion(), version.GitCommitMessage, version.GitCommitSHA)

	readConfig := types.ReadConfig{}
	osEnv := types.OsEnv{}
	cfg := readConfig.Read(osEnv)


	var proxyFunc = proxy.NewHandlerFunc(cfg.ReadTimeout,
		handlers.NewFunctionLookup(routing.NewDefaultProviderRouting(cfg.Providers, cfg.DefaultProvider)))

	bootstrapHandlers := bootTypes.FaaSHandlers{
		FunctionProxy:  proxyFunc,
		DeleteHandler:  proxyFunc,
		DeployHandler:  proxyFunc,
		FunctionReader: handlers.MakeFunctionReader(),
		ReplicaReader:  handlers.MakeReplicaReader(),
		ReplicaUpdater: handlers.MakeReplicaUpdater(),
		UpdateHandler:  proxyFunc,
		HealthHandler:  handlers.MakeHealthHandler(),
		InfoHandler:    handlers.MakeInfoHandler(version.BuildVersion(), version.GitCommitSHA),
	}

	bootstrapConfig := bootTypes.FaaSConfig{
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		TCPPort:         &cfg.Port,
		EnableHealth:    true,
		EnableBasicAuth: false,
	}

	log.Infof("listening on port %d", cfg.Port)
	bootstrap.Serve(&bootstrapHandlers, &bootstrapConfig)
}
