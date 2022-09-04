package yamler

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"

	"gopkg.in/yaml.v2"
)

type Conf struct {
	Description string                   `yaml:"description,omitempty"`
	Name        string                   `yaml:"name,omitempty"`
	Services    []string                 `yaml:"services,omitempty"`
	Properties  []map[string]interface{} `yaml:"properties,omitempty"`
	Return      []string                 `yaml:"return,omitempty"`
	Enabled     bool                     `yaml:"enabled,omitempty"`
	Find        Find                     `yaml:"find,omitempty"`
}

type Find struct {
	Who    []string
	To     []string
	With   []string            `yaml:"with,omitempty"`
	Target []map[string]string `yaml:"target,omitempty"`
}

func walkMatch(root, pattern string) []string {
	var a []string
	err := filepath.WalkDir(root, func(s string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if ok, err := regexp.Match(pattern, []byte(filepath.Ext(d.Name()))); ok {
			a = append(a, s)
		} else {
			return err
		}
		return nil
	})
	if err != nil {
		log.Fatalln(err)
	}
	return a
}

func normalizePath(path string) string {
	usr, _ := user.Current()
	dir := usr.HomeDir
	if path == "~" {
		path = dir
	} else if strings.HasPrefix(path, "~/") {
		path = filepath.Join(dir, path[2:])
	}

	path, _ = filepath.Abs(filepath.Clean(path))
	return path
}

func GetFiles(basepath string) []string {
	return walkMatch(normalizePath("./assess/rules/"), ".ya?ml")
}

func (c *Conf) GetConf(file string) *Conf {
	yamlFile, err := ioutil.ReadFile(normalizePath(file))
	if err != nil {
		log.Printf("yamlFile.Get err #%v", err)
	}
	c.Enabled = true // Default value is: Enabled
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}
