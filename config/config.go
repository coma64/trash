package config

import "time"

type Globals struct {
	DbPath string `default:"./trash.db" type:"path"`
	Debug  bool   `default:"false"`
}

var Config struct {
	Serve struct {
		Globals
		RawServerAddress  string `default:"127.0.0.1:2222"`
		HttpServerAddress string `default:"127.0.0.1:6969"`

		SnippetRetentionTime             time.Duration `default:"168h"`     // 1 week
		SnippetGarbageCollectionInterval time.Duration `default:"12h"`      // 1 hour
		SnippetMaxSizeBytes              uint          `default:"10485760"` // 10MB

		AutomaticallyApplyMigrations bool `default:"true"`
	} `cmd:""`
}
