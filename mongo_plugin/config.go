package main

import (
	"unsafe"

	"github.com/fluent/fluent-bit-go/output"
)

type Config struct {
	URL      string
	Database string
}

func GetURL(ctx unsafe.Pointer) string {
	return output.FLBPluginConfigKey(ctx, "url")
}

func GetDatabase(ctx unsafe.Pointer) string {
	return output.FLBPluginConfigKey(ctx, "database")
}

func GetConfig(ctx unsafe.Pointer) *Config {
	return &Config{
		URL:      GetURL(ctx),
		Database: GetDatabase(ctx),
	}
}
