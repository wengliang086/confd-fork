package template

import (
	"confd-fork/backends"
	"confd-fork/log"
	"confd-fork/util"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/kelseyhightower/memkv"
	"path/filepath"
	"strconv"
	"strings"

	//"github.com/xordataexchange/crypt/encoding/secconf"
	"os"
)

type Config struct {
	//ConfDir       string `toml:"confdir"`
	ConfigDir     string
	KeepStageFile bool
	Noop          bool   `toml:"noop"`
	Prefix        string `toml:"prefix"`
	StoreClient   backends.StoreClient
	SyncOnly      bool `toml:"sync-only"`
	TemplateDir   string
	PGPPrivateKey []byte
}

// TemplateResourceConfig holds the parsed template resource.
type TemplateResourceConfig struct {
	TemplateResource TemplateResource `toml:"template"`
}

// TemplateResource is the representation of a parsed template resource.
type TemplateResource struct {
	CheckCmd      string `toml:"check_cmd"`
	Dest          string
	FileMode      os.FileMode
	Gid           int
	Keys          []string
	Mode          string
	Prefix        string
	ReloadCmd     string `toml:"reload_cmd"`
	Src           string
	StageFile     *os.File
	Uid           int
	funcMap       map[string]interface{}
	lastIndex     uint64
	keepStageFile bool
	noop          bool
	store         memkv.Store
	storeClient   backends.StoreClient
	syncOnly      bool
	PGPPrivateKey []byte
}

var ErrEmptySrc = errors.New("empty src template")

// NewTemplateResource creates a TemplateResource.
func NewTemplateResource(path string, config Config) (*TemplateResource, error) {
	if config.StoreClient == nil {
		return nil, errors.New("A valid StoreClient is required.")
	}

	// Set the default uid and gid so we can determine if it was
	// unset from configuration.
	tc := &TemplateResourceConfig{TemplateResource{Uid: -1, Gid: -1}}

	log.Debug("Loading template resource from " + path)
	_, err := toml.DecodeFile(path, &tc)
	if err != nil {
		return nil, fmt.Errorf("Cannot process template resource %s - %s", path, err.Error())
	}

	tr := tc.TemplateResource
	tr.keepStageFile = config.KeepStageFile
	tr.noop = config.Noop
	tr.storeClient = config.StoreClient
	tr.funcMap = newFuncMap()
	tr.store = memkv.New()
	tr.syncOnly = config.SyncOnly
	addFuncs(tr.funcMap, tr.store.FuncMap)

	if config.Prefix != "" {
		tr.Prefix = config.Prefix
	}

	if !strings.HasPrefix(tr.Prefix, "/") {
		tr.Prefix = "/" + tr.Prefix
	}

	if len(config.PGPPrivateKey) > 0 {
		tr.PGPPrivateKey = config.PGPPrivateKey
		addCryptFuncs(&tr)
	}

	if tr.Src == "" {
		return nil, ErrEmptySrc
	}

	if tr.Uid == -1 {
		tr.Uid = os.Geteuid()
	}

	if tr.Gid == -1 {
		tr.Gid = os.Getegid()
	}

	tr.Src = filepath.Join(config.TemplateDir, tr.Src)
	return &tr, nil
}

func addCryptFuncs(tr *TemplateResource) {
	addFuncs(tr.funcMap, map[string]interface{}{
		"cget": func(key string) (memkv.KVPair, error) {
			kv, err := tr.funcMap["get"].(func(string) (memkv.KVPair, error))(key)
			if err == nil {
				//var b []byte
				//b, err = secconf.Decode([]byte(kv.Value), bytes.NewBuffer(tr.PGPPrivateKey))
				//if err == nil {
				//	kv.Value = string(b)
				//}
			}
			return kv, err
		},
	})
}

func printContent(t *TemplateResource) error {
	result, err := t.storeClient.GetValues(util.AppendPrefix(t.Prefix, t.Keys))
	if err != nil {
		return err
	}

	log.Debug("Got the following map from store: %v", result)
	return nil
}

func (t *TemplateResource) process() error {
	log.Debug("Test: sync file")

	if err := t.setFileMode(); err != nil {
		return err
	}

	if err := printContent(t); err != nil {
		return err
	}

	//if err := t.setVars(); err != nil {
	//	return err
	//}
	//if err := t.createStageFile(); err != nil {
	//	return err
	//}
	//if err := t.sync(); err != nil {
	//	return err
	//}
	return nil
}

func (t *TemplateResource) setFileMode() error {
	if t.Mode == "" {
		if !util.IsFileExist(t.Dest) {
			t.FileMode = 0644
		} else {
			fi, err := os.Stat(t.Dest)
			if err != nil {
				return err
			}
			t.FileMode = fi.Mode()
		}
	} else {
		mode, err := strconv.ParseUint(t.Mode, 0, 32)
		if err != nil {
			return err
		}
		t.FileMode = os.FileMode(mode)
	}
	return nil
}
