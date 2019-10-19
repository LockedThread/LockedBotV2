package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Configuration struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"database-name"`
	Tables   struct {
		UserTable      string `yaml:"users"`
		ResourcesTable string `yaml:"resources"`
	}
}

func (c *Configuration) SetupConfig() *Configuration {
	yamlFile, err := ioutil.ReadFile("config.yml")
	checkErr(err)
	err = yaml.Unmarshal(yamlFile, &c)
	checkErr(err)
	return c
}
