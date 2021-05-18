package redis

import (
	"confd-fork/log"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

type watchResponse struct {
	waitIndex uint64
	err       error
}

type Client struct {
	client    redis.Conn
	machines  []string
	password  string
	separator string
}

// Iterate through `machines`, trying to connect to each in turn.
// Returns the first successful connection or the last error encountered.
// Assumes that `machines` is non-empty.
func tryConnect(machines []string, password string, timeout bool) (redis.Conn, int, error) {
	var err error
	for _, address := range machines {
		var conn redis.Conn
		var db int

		idx := strings.Index(address, "/")
		if idx != -1 {
			db, err = strconv.Atoi(address[idx+1:])
			if err == nil {
				address = address[:idx]
			}
		}

		network := "tcp"
		if _, err = os.Stat(address); err == nil {
			network = "unix"
		}
		log.Debug(fmt.Sprintf("Trying to connect to redis node %s", address))

		var dialOps []redis.DialOption
		if timeout {
			dialOps = []redis.DialOption{
				redis.DialConnectTimeout(time.Second),
				redis.DialReadTimeout(time.Second * 2),
				redis.DialWriteTimeout(time.Second * 3),
				redis.DialDatabase(db),
			}
		} else {
			dialOps = []redis.DialOption{
				redis.DialConnectTimeout(time.Second),
				redis.DialWriteTimeout(time.Second * 3),
				redis.DialDatabase(db),
			}
		}

		if password != "" {
			dialOps = append(dialOps, redis.DialPassword(password))
		}

		conn, err = redis.Dial(network, address, dialOps...)
		if err != nil {
			continue
		}

		return conn, db, nil
	}

	return nil, 0, err
}

func NewRedisClient(machines []string, password string, separator string) (*Client, error) {
	if separator == "" {
		separator = "/"
	}
	log.Debug(fmt.Sprintf("Redis Separator: %#v", separator))

	var err error
	clientWrapper := &Client{}
	clientWrapper.client, _, err = tryConnect(machines, password, true)
	return clientWrapper, err
}

func (c *Client) transform(key string) string {
	if c.separator == "/" {
		return key
	}
	k := strings.TrimPrefix(key, "/")
	return strings.Replace(k, "/", c.separator, -1)
}

func (c *Client) clean(key string) string {
	k := key
	if !strings.HasPrefix(k, "/") {
		k = "/" + k
	}
	return strings.Replace(k, c.separator, "/", -1)
}

func (c Client) GetValues(keys []string) (map[string]string, error) {
	panic("implement me")
}

func (c Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	panic("implement me")
}
