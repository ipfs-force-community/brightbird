package main

import "github.com/ipfs-force-community/brightbird/env"

type Config struct {
	NameSpace   string
	PluginStore string

	Timeout int
	Listen  string

	MongoURL string
	DBName   string

	Mysql string

	TaskId   string
	Registry string

	GlobalParams env.GlobalParams
}
