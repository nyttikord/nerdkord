package main

import (
	"github.com/anhgelus/gokord"
	"github.com/pelletier/go-toml/v2"
	"gorm.io/gorm"
)

type Config struct {
	Debug    bool   `toml:"debug"`
	Author   string `toml:"author"`
	Database string `toml:"database"`
}

func (c *Config) Connect() (*gorm.DB, error) {
	return gorm.Open(nil, &gorm.Config{})
}

func (c *Config) IsDebug() bool {
	return c.Debug
}

func (c *Config) GetAuthor() string {
	return c.Author
}

func (c *Config) GetRedisCredentials() *gokord.RedisCredentials {
	return nil
}

func (c *Config) GetSQLCredentials() gokord.SQLCredentials {
	return c
}

func (c *Config) SetDefaultValues() {
	c.Debug = false
	c.Author = "nyttikord"
	c.Database = "nerdkord.db"
}

func (c *Config) Marshal() ([]byte, error) {
	return toml.Marshal(c)
}

func (c *Config) Unmarshal(bytes []byte) error {
	return toml.Unmarshal(bytes, c)
}
