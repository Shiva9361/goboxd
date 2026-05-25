package internal

import (
	"log"

	"github.com/spf13/viper"
)

type Resource struct {
	WallTime   int `mapstructure:"wall_time_s"`
	Memory     int `mapstructure:"memory_kb"`
	MaxProcess int `mapstructure:"max_process"`
}

type Options struct {
	Cmd    string   `mapstructure:"cmd"`
	Args   []string `mapstructure:"args"`
	Limits Resource `mapstructure:"limits"`
	Flags  []string `mapstructure:"flag_allowlist"`
}

type CodeConfig struct {
	ID              string  `mapstructure:"id"`
	Name            string  `mapstructure:"name"`
	Source_filename string  `mapstructure:"source"`
	Artifact        string  `mapstructure:"artifact"`
	Build_options   Options `mapstructure:"build"`
	Run_options     Options `mapstructure:"run"`
}

type CodeRunner struct {
	Configs   []CodeConfig
	ConfigMap map[string]int // felt this is cleaner
}

// Initialize is used to load up all the config yaml
// to support all required languages
func (c *CodeRunner) Initialize() {

	viper.SetConfigName("settings")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")

	if err := viper.ReadInConfig(); err != nil {
		log.Panicln("Reading config file failed with: ", err)
	}

	if err := viper.UnmarshalKey("languages", &c.Configs); err != nil {
		log.Panicln("Failed to parse config file with: ", err)
	}

	c.ConfigMap = make(map[string]int)

	for i, config := range c.Configs {
		c.ConfigMap[config.ID] = i
	}

}
