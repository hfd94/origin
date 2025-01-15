package config

import (
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary
var config = map[string]*Config{}

const Cluster = "cluster"

func init() {
	NewConfig(Cluster, "./bin/cluster")
}

type WithConfig func(src map[string]interface{})

type Config struct {
	path  string
	value map[string]interface{}
}

func GetConfig(key string) *Config {
	return config[key]
}

func NewConfig(key string, path string) *Config {
	if config[path] != nil {
		return config[path]
	}
	config[key] = &Config{path: path}
	return config[key]
}

func (c *Config) Load(opts ...WithConfig) error {
	opt := func(src map[string]interface{}) {}
	if len(opts) > 0 {
		opt = opts[0]
	}

	// 获取文件/目录的信息
	info, err := os.Stat(c.path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("路径不存在: %s", c.path)
		}
		return fmt.Errorf("获取路径信息失败: %s", err)
	}

	if info.IsDir() {
		err := filepath.Walk(c.path, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			ext := filepath.Ext(p)
			// 如果是文件且扩展名是 json 或 yaml
			if !info.IsDir() && (ext == ".json" || ext == ".yaml" || ext == ".yml") {
				d, err := os.ReadFile(p)
				if err != nil {
					return fmt.Errorf("加载文件失败: %s, 错误: %v", p, err)
				}
				if ext == ".json" {
					cnf, err := loadJSON(d)
					if err != nil {
						return fmt.Errorf("解析配置失败: %s, 错误: %v", p, err)
					}
					opt(cnf)
				} else if ext == ".yaml" || ext == ".yml" {
					cnf, err := loadYAML(d)
					if err != nil {
						return fmt.Errorf("解析配置失败: %s, 错误: %v", p, err)
					}
					opt(cnf)
				}
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("遍历目录时发生错误: %v", err)
		}
	} else {
		d, err := os.ReadFile(c.path)
		if err != nil {
			return fmt.Errorf("加载文件失败: %s, 错误: %v", c.path, err)
		}
		ext := filepath.Ext(c.path)
		if ext == ".json" {
			cnf, err := loadJSON(d)
			if err != nil {
				return fmt.Errorf("解析配置失败: %s, 错误: %v", c.path, err)
			}
			opt(cnf)
		} else if ext == ".yaml" || ext == ".yml" {
			cnf, err := loadYAML(d)
			if err != nil {
				return fmt.Errorf("解析配置失败: %s, 错误: %v", c.path, err)
			}
			opt(cnf)
		}
	}
	return nil
}

func (c *Config) Value() map[string]interface{} {
	return c.value
}

func loadJSON(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func loadYAML(data []byte) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := yaml.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
