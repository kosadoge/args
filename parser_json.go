package args

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

type JsonParser struct {
	path *string
}

func (j *JsonParser) Parse(set func(name, value string) error) error {
	if j.path == nil || *j.path == "" {
		return nil
	}

	f, err := os.Open(*j.path)
	if err != nil {
		return err
	}
	defer f.Close()

	d := json.NewDecoder(f)
	d.UseNumber()

	var obj map[string]any
	if err := d.Decode(&obj); err != nil {
		return fmt.Errorf("decode json config file failed: %w", err)
	}

	if err := processJsonObject("", obj, set); err != nil {
		return err
	}

	return nil
}

func processJsonObject(prevKey string, obj map[string]any, set func(name, value string) error) error {
	for k, v := range obj {
		if prevKey != "" {
			k = prevKey + "." + k
		}
		switch v := v.(type) {
		case map[string]any:
			if err := processJsonObject(k, v, set); err != nil {
				return err
			}
		case []any:
			for _, v := range v {
				s, err := jsonValuetoString(v)
				if err != nil {
					return err
				}
				if err := set(k, s); err != nil {
					return err
				}
			}
		default:
			s, err := jsonValuetoString(v)
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

func jsonValuetoString(v any) (string, error) {
	switch v := v.(type) {
	case string:
		return v, nil
	case json.Number:
		return v.String(), nil
	case bool:
		return strconv.FormatBool(v), nil
	default:
		return "", fmt.Errorf("unexpect json type: %#v", v)
	}
}
