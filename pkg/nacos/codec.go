package nacos

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

func UnmarshalYamlToMap(yamlStr string) (map[interface{}]interface{}, error) {
	var m map[interface{}]interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
		return nil, err
	}
	return m, nil
}

func MergeMap(dest, src map[interface{}]interface{}) map[interface{}]interface{} {
	out := make(map[interface{}]interface{}, len(dest))
	for k, v := range dest {
		out[k] = v
	}

	for k, v := range src {
		value := v
		if av, ok := out[k]; ok {
			if v, ok := v.(map[interface{}]interface{}); ok {
				if av, ok := av.(map[interface{}]interface{}); ok {
					out[k] = MergeMap(av, v)
				} else {
					out[k] = v
				}
			} else {
				out[k] = value
			}
		} else {
			out[k] = v
		}
	}
	return out
}

func MarshalObjectToYamlString(obj interface{}) (string, error) {
	bs, err := yaml.Marshal(obj)
	if err != nil {
		return "", fmt.Errorf("marshal obj [%#v] faild, err: %w", obj, err)
	}
	return string(bs), nil
}
