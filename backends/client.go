package backends

import (
	"confd-fork/backends/etcdv3"
	"confd-fork/backends/redis"
	"errors"
)

type StoreClient interface {
	GetValues(keys []string) (map[string]string, error)
	WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error)
}

func New(config Config) (StoreClient, error) {
	if config.Backend == "" {
		config.Backend = "etcdv3"
	}

	backendNodes := config.BackendNodes

	switch config.Backend {
	case "etcdv3":
		return etcdv3.NewEtcdClient(backendNodes, config.ClientCert, config.ClientKey, config.ClientCaKeys,
			config.BasicAuth, config.Username, config.Password)
	case "redis":
		return redis.NewRedisClient(backendNodes, config.ClientKey, config.Separator)
	}

	return nil, errors.New("invalid backend")
}
