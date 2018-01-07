package file

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/kelseyhightower/confd/log"
	"gitlab.com/pastdev/s2i/clconf/clconf"
)

// Client provides a shell for the yaml client
type Client struct {
	yamlFiles []string
}

func NewClconfClient(yamlFiles string) (*Client, error) {
	return &Client{clconf.Splitter.Split(yamlFiles, -1)}, nil
}

func (c *Client) GetValues(keys []string) (map[string]string, error) {
	vars := make(map[string]string)
	yamlMap, err := clconf.LoadConf(c.yamlFiles, []string{})
	if err != nil {
		return vars, err
	}
	nodeWalk(yamlMap, "", vars)
	log.Debug(fmt.Sprintf("Key Map: %#v", vars))

	return vars, nil
}

// nodeWalk recursively descends nodes, updating vars.
func nodeWalk(node map[interface{}]interface{}, key string, vars map[string]string) error {
	for k, v := range node {
		key := key + "/" + k.(string)

		switch v.(type) {
		case map[interface{}]interface{}:
			nodeWalk(v.(map[interface{}]interface{}), key, vars)
		case []interface{}:
			for _, j := range v.([]interface{}) {
				switch j.(type) {
				case map[interface{}]interface{}:
					nodeWalk(j.(map[interface{}]interface{}), key, vars)
				default:
					vars[fmt.Sprintf("%s/%v", key, j)] = ""
				}
			}
		default:
			vars[key] = fmt.Sprintf("%v", v)
		}

	}
	return nil
}

func (c *Client) WatchPrefix(prefix string, keys []string, waitIndex uint64, stopChan chan bool) (uint64, error) {
	if waitIndex == 0 {
		return 1, nil
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return 0, err
	}
	defer watcher.Close()

	for _, filepath := range c.yamlFiles {
		err = watcher.Add(filepath)
		if err != nil {
			return 0, err
		}
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Remove == fsnotify.Remove {
				return 1, nil
			}
		case err := <-watcher.Errors:
			return 0, err
		case <-stopChan:
			return 0, nil
		}
	}
	return waitIndex, nil
}
