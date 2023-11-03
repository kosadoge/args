package args

import (
	"fmt"
	"io"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

type YamlParser struct {
	path *string
}

func (y *YamlParser) Parse(set func(name, value string) error) error {
	if y.path == nil || *y.path == "" {
		return nil
	}

	f, err := os.Open(*y.path)
	if err != nil {
		return err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)

	var obj map[string]any
	if err := d.Decode(&obj); err != nil && err != io.EOF {
		return fmt.Errorf("decode yaml config file failed: %w", err)
	}

	if err := processYamlObject("", obj, set); err != nil {
		return err
	}

	return nil
}

func processYamlObject(prevKey string, obj map[string]any, set func(name, value string) error) error {
	for k, v := range obj {
		if prevKey != "" {
			k = prevKey + "." + k
		}
		switch v := v.(type) {
		case map[string]any:
			if err := processYamlObject(k, v, set); err != nil {
				return err
			}
		case []any:
			for _, v := range v {
				s, err := yamlValueToString(v)
				if err != nil {
					return err
				}
				if err := set(k, s); err != nil {
					return err
				}
			}
		default:
			s, err := yamlValueToString(v)
			if err != nil {
				return err
			}
			if err := set(k, s); err != nil {
				return err
			}
		}
	}
	return nil
}

func yamlValueToString(v any) (string, error) {
	switch v := v.(type) {
	case byte:
		return string([]byte{v}), nil
	case string:
		return v, nil
	case bool:
		return strconv.FormatBool(v), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case int:
		return strconv.Itoa(v), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case float64:
		return strconv.FormatFloat(v, 'g', -1, 64), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("unexpect yaml type: %#v", v)
	}
}
