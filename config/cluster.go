package config

import (
	"errors"
)
import "github.com/go-viper/mapstructure/v2"

func SetClusterPath(path string) { GetConfig(Cluster).path = path }

func deepMerge(dest, src map[string]interface{}) {
	for key, srcVal := range src {

		destVal, exists := dest[key]
		if !exists {
			dest[key] = srcVal
			continue
		}

		if srcMap, ok := srcVal.(map[string]interface{}); ok {
			if destMap, ok := destVal.(map[string]interface{}); ok {
				deepMerge(destMap, srcMap)
			} else {
				dest[key] = srcVal
			}
		} else if srcSlice, ok := srcVal.([]interface{}); ok {
			if destSlice, ok := destVal.([]interface{}); ok {
				dest[key] = append(destSlice, srcSlice...)
			} else {
				dest[key] = srcVal
			}
		} else {
			dest[key] = srcVal
		}
	}
}

func ClusterLoad() error {
	a := GetConfig(Cluster).Load(func(src map[string]interface{}) {
		deepMerge(GetConfig(Cluster).value, src)
	})
	return a
}

func GetSystemConfigParse(key string, res interface{}) error {
	if m, ok := GetConfig(Cluster).value[key]; ok {
		err := mapstructure.Decode(m, res)
		return err
	}
	return errors.New("key not found")
}

func GetSystemConfigKey(key string) interface{} {
	return GetConfig(Cluster).value[key]
}

func GetSystemConfig() map[string]interface{} {
	return GetConfig(Cluster).value
}
