package main

import (
	"confd-fork/backends"
	"confd-fork/log"
	"confd-fork/resource/template"
	"flag"
	"github.com/BurntSushi/toml"
	"io/ioutil"
	"os"
)

type TemplateConfig = template.Config
type BackendsConfig = backends.Config

type Config struct {
	TemplateConfig
	BackendsConfig
	Interval      int    `toml:"interval"`
	SecretKeyring string `toml:"secret_keyring"`
	SRVDomain     string `toml:"srv_domain"`
	SRVRecord     string `toml:"srv_record"`
	LogLevel      string `toml:"log-level"`
	Watch         bool   `toml:"watch"`
	PrintVersion  bool
	ConfigFile    string
	OneTime       bool
}

var config Config

func init() {
	flag.StringVar(&config.AuthToken, "auth-token", "", "Auth bearer token to use")

	flag.BoolVar(&config.Watch, "watch", true, "enable watch support")
	flag.BoolVar(&config.OneTime, "onetime", false, "")
	flag.StringVar(&config.Backend, "backend", "etcdv3", "")
	flag.Var(&config.BackendNodes, "node", "list of backend nodes")
	flag.StringVar(&config.ConfigDir, "confdir", "/Users/phoenix/Workspace/dragon-server-code/scripts/confd", "list of backend nodes")
}

func initConfig() error {
	_, err := os.Stat(config.ConfigFile)
	if os.IsNotExist(err) {
		log.Debug("Skipping config file.")
	} else {
		log.Debug("Loading " + config.ConfigFile)
		configBytes, err := ioutil.ReadFile(config.ConfigFile)
		if err != nil {
			return err
		}

		_, err = toml.Decode(string(configBytes), &config)
		if err != nil {
			return err
		}
	}

	return nil
}
