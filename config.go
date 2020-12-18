package main

import (
	"flag"
	"path/filepath"

	"golang.org/x/xerrors"

	"github.com/BurntSushi/toml"
)

var (
	// オプションを指定したときとしていないときの区別が難しいのですべて String で受ける
	optConfigFile     = flag.String("config", "./impas.toml", "config file name which includes dependency rules")
	optIgnoreExternal = flag.String("ignoreExternal", "", "ignore imported packages NOT included in the Root project if true")
	optRecursive      = flag.String("recursive", "", "search imported packages recursively if true")
	optConcurrency    = flag.Int64("concurrency", 1, "number of concurrency")
)

type Constraint struct {
	From  string
	Allow []string
}

type Config struct {
	AbsPath        string
	Constraint     []Constraint
	IgnoreExternal bool
	Recursive      bool
	Concurrency    int64
}

func (c *Config) MaxDepth() int {
	if c.Recursive {
		return 0
	}
	return 1
}

func NewConfig() (*Config, error) {
	flag.Parse()

	c, err := loadTomlConfig(*optConfigFile)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	applyCommandLineOptions(c)
	return c, nil
}

func applyCommandLineOptions(c *Config) {
	if *optIgnoreExternal != "" {
		if *optIgnoreExternal == "false" {
			c.IgnoreExternal = false
		} else {
			c.IgnoreExternal = true
		}
	}
	if *optRecursive != "" {
		if *optRecursive == "false" {
			c.Recursive = false
		} else {
			c.Recursive = true
		}
	}
	c.Concurrency = *optConcurrency
}

func loadTomlConfig(fileName string) (*Config, error) {
	var config Config
	_, err := toml.DecodeFile(fileName, &config)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	absPath, err := filepath.Abs(fileName)
	if err != nil {
		return nil, xerrors.Errorf(": %w", err)
	}
	config.AbsPath = filepath.Clean(absPath)
	return &config, nil
}
