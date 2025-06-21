package main

import (
	"fmt"
	"github.com/anhgelus/gokord"
	"github.com/glebarez/sqlite"
	"github.com/pelletier/go-toml/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	Debug       bool            `toml:"debug"`
	Author      string          `toml:"author"`
	UsePostgres bool            `toml:"use_postgres_instead_of_sqlite"`
	SQLite      *SQLiteConfig   `toml:"sqlite"`
	Postgres    *PostgresConfig `toml:"postgres"`
}

type SQLiteConfig struct {
	Path string `toml:"path"`
}

func (s *SQLiteConfig) Connect() (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(s.Path), &gorm.Config{})
}

func (s *SQLiteConfig) SetDefaultValues() {
	s.Path = "nerdkord.db"
}

type PostgresConfig struct {
	Host     string `toml:"host"`
	User     string `toml:"user"`
	Password string `toml:"password"`
	DBName   string `toml:"db_name"`
	Port     int    `toml:"port"`
	TimeZone string `toml:"time_zone"`
}

func (p *PostgresConfig) generateDsn() string {
	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable TimeZone=%s",
		p.Host, p.User, p.Password, p.DBName, p.Port, p.TimeZone,
	)
}

func (p *PostgresConfig) Connect() (*gorm.DB, error) {
	return gorm.Open(postgres.Open(p.generateDsn()), &gorm.Config{})
}

func (p *PostgresConfig) SetDefaultValues() {
	p.Host = "localhost"
	p.User = "nerdkord"
	p.Password = "password"
	p.DBName = "nerdkord"
	p.Port = 5432
	p.TimeZone = "Europe/Paris"
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
	if c.UsePostgres {
		return c.Postgres
	}
	return c.SQLite
}

func (c *Config) SetDefaultValues() {
	c.Debug = false
	c.Author = "nyttikord"
	c.UsePostgres = false

	c.Postgres = &PostgresConfig{}
	c.Postgres.SetDefaultValues()

	c.SQLite = &SQLiteConfig{}
	c.SQLite.SetDefaultValues()
}

func (c *Config) Marshal() ([]byte, error) {
	return toml.Marshal(c)
}

func (c *Config) Unmarshal(bytes []byte) error {
	return toml.Unmarshal(bytes, c)
}
