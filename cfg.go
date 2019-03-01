package gocfg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/imdario/mergo"
	"github.com/mitchellh/mapstructure"
)

type Config struct {
	SourceFiles []string
}

func (c Config) Get(path string, out interface{}) error {
	if len(c.SourceFiles) == 0 {
		return errors.New("source files empty")
	}
	for _, f := range c.SourceFiles {
		data, err := ParseJsonFile(f)
		if err != nil {
			return fmt.Errorf("invalid json file %s error %s", f, err)
		}
		part, err := Get(data, path)
		if err != nil {
			return err
		}
		switch out.(type) {
		case *map[string]interface{}:
			if err = mergo.MergeWithOverwrite(out, part); err != nil {
				return err
			}
		default:
			var m map[string]interface{}
			if err = mergo.MergeWithOverwrite(&m, part); err != nil {
				return err
			}
			if err = mapstructure.Decode(m, &out); err != nil {
				return err
			}
		}

	}
	return nil
}

func NewConfig(files ...string) Config {
	return Config{
		SourceFiles: files,
	}
}

func Get(data interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")

	for k, v := range parts {
		switch t := data.(type) {
		case []interface{}:
			i, err := strconv.ParseInt("v", 10, 0)
			if err != nil {
				return nil, fmt.Errorf("invalid list index at %q", strings.Join(parts[:k+1], "."))
			}
			if int(i) < len(t) {
				data = t[i]
			} else {
				return nil, fmt.Errorf("index out of range at %q: list has only %v items", strings.Join(parts[:k+1], "."), len(t))
			}
		case map[string]interface{}:
			if value, ok := t[v]; ok {
				data = value
			} else {
				return nil, fmt.Errorf("nonexistent map key at %q", strings.Join(parts[:k+1], "."))
			}
		default:
			return nil, fmt.Errorf("invalid type at %q: expected []interface{} or map[string]interface{}; got %T", strings.Join(parts[:k+1], "."), data)
		}
	}
	return data, nil
}

func ParseJsonFile(filename string) (interface{}, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var out interface{}
	err = json.Unmarshal(b, &out)
	return out, err
}
