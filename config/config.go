// Created on 2015-07-20
// summary: config
// author: YangHaitao

package config

import (
    "os"
    "fmt"
    "go-gypsy/yaml"
)

var Config *yaml.File

func GetConfig(config_path string) *yaml.File {
    if Config != nil {
        return Config
    }

    if _, err := os.Stat(config_path); os.IsNotExist(err) {
        fmt.Printf("Init Config (%s) error: config_path does not exist!\n", config_path)
        return Config
    }

    tmp_config, err := yaml.ReadFile(config_path)
    if err != nil {
        fmt.Printf("Init Config (%s) error: %s\n", config_path, err)
        return Config
    }
    Config = tmp_config
    return Config
}
