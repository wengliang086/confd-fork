package template

import (
	"confd-fork/log"
	"confd-fork/util"
	"fmt"
	"sync"
	"time"
)

type Processor interface {
	Process()
}

func Process(config Config) error {
	templateResources, err := getTemplateResources(config)
	if err != nil {
		return err
	}
	return process(templateResources)
}

func process(ts []*TemplateResource) error {
	var lastErr error
	for _, t := range ts {
		if err := t.process(); err != nil {
			log.Error(err.Error())
			lastErr = err
		}
	}
	return lastErr
}

type intervalProcessor struct {
	config   Config
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	interval int
}

func (p *intervalProcessor) Process() {

}

func IntervalProcessor(config Config, stopChan, doneChan chan bool, errChan chan error, interval int) Processor {
	return &intervalProcessor{config, stopChan, doneChan, errChan, interval}
}

type watchProcessor struct {
	config   Config
	stopChan chan bool
	doneChan chan bool
	errChan  chan error
	wg       sync.WaitGroup
}

func (p *watchProcessor) Process() {
	defer close(p.doneChan)
	templates, err := getTemplateResources(p.config)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	for _, template := range templates {
		t := template
		p.wg.Add(1)
		go p.monitorPrefix(t)
	}
	p.wg.Wait()
}

func (p *watchProcessor) monitorPrefix(t *TemplateResource) {
	defer p.wg.Done()
	keys := util.AppendPrefix(t.Prefix, t.Keys)
	for {
		idx, err := t.storeClient.WatchPrefix(t.Prefix, keys, t.lastIndex, p.stopChan)
		if err != nil {
			p.errChan <- err

			time.Sleep(time.Second)
			continue
		}

		t.lastIndex = idx
		if err := t.process(); err != nil {
			p.errChan <- err
		}
	}
}

func WatchProcessor(config Config, stopChan, doneChan chan bool, errChan chan error) Processor {
	var wg sync.WaitGroup
	return &watchProcessor{config, stopChan, doneChan, errChan, wg}
}

func getTemplateResources(config Config) ([]*TemplateResource, error) {
	var lastError error
	templates := make([]*TemplateResource, 0)

	log.Debug("Loading template resources from confdir " + config.ConfigDir)
	if !util.IsFileExist(config.ConfigDir) {
		log.Warning(fmt.Sprintf("Cannot load template resources: confdir '%s' does not exist", config.ConfigDir))
		return nil, nil
	}

	paths, err := util.RecursiveFilesLookup(config.ConfigDir, "*toml")
	if err != nil {
		return nil, err
	}

	if len(paths) < 1 {
		log.Warning("Found no templates")
	}

	for _, path := range paths {
		log.Debug(fmt.Sprintf("Found template: %s", path))
		resource, err := NewTemplateResource(path, config)
		if err != nil {
			lastError = err
			continue
		}
		templates = append(templates, resource)
	}

	return templates, lastError
}
