package config

import (
	"io/ioutil"
	"log"
	"regexp"

	"gopkg.in/yaml.v2"

	"github.com/tsiemens/gmail-tools/util"
)

const (
	ConfigYamlFileName = "config.yaml"

	caseIgnore = "(?i)"
)

type Config struct {
	InterestingMessageQuery    string            `yaml:"InterestingMessageQuery"`
	UninterestingLabelPatterns []string          `yaml:"UninterestingLabelPatterns"`
	InterestingLabelPatterns   []string          `yaml:"InterestingLabelPatterns"`
	ApplyLabelToUninteresting  string            `yaml:"ApplyLabelToUninteresting"`
	ApplyLabelOnTouch          string            `yaml:"ApplyLabelOnTouch"`
	LabelColors                map[string]string `yaml:"LabelColors"`

	UninterLabelRegexps []*regexp.Regexp
	InterLabelRegexps   []*regexp.Regexp
	ConfigFile          string
}

func loadConfig() *Config {
	confFname := util.RequiredHomeDirAndFile(util.UserAppDirName, ConfigYamlFileName)

	confData, err := ioutil.ReadFile(confFname)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	conf := &Config{}
	conf.ConfigFile = confFname
	err = yaml.Unmarshal(confData, conf)
	if err != nil {
		log.Fatalf("Could not unmarshal: %v", err)
	}
	util.Debugf("config: %+v\n", conf)

	for _, pat := range conf.UninterestingLabelPatterns {
		re, err := regexp.Compile(caseIgnore + pat)
		if err != nil {
			break
		}
		conf.UninterLabelRegexps = append(conf.UninterLabelRegexps, re)
	}
	if err == nil {
		for _, pat := range conf.InterestingLabelPatterns {
			re, err := regexp.Compile(caseIgnore + pat)
			if err != nil {
				break
			}
			conf.InterLabelRegexps = append(conf.InterLabelRegexps, re)
		}
	}
	if err != nil {
		log.Fatalf("Failed to load config: \"%s\"", err)
	}
	return conf
}

var appConfig *Config

func AppConfig() *Config {
	if appConfig == nil {
		appConfig = loadConfig()
	}
	return appConfig
}