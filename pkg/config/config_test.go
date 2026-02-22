package config_test

import (
	"testing"

	config "mox/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "config not found", r)
		}

	}()

	config.NewDefaultConfig()
}

func TestConfigWithPath(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Log(r)
			assert.Fail(t, "panic triggered")
		}
	}()

	config.NewConfig(config.ConfigParam{
		ConfigName: "config",
		ConfigType: "toml",
		Path:       "./testdata",
	})
}

func TestConfigValidateTriggered(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			assert.Fail(t, "panic not triggered")
		}
	}()

	config.NewConfig(config.ConfigParam{
		ConfigName: "config_failed",
		ConfigType: "toml",
		Path:       "./testdata",
	})
}

func TestConfigArray(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Fail(t, "panic triggered")
		}
	}()

	c := config.NewConfig(config.ConfigParam{
		ConfigName: "config",
		ConfigType: "toml",
		Path:       "./testdata",
	})

	assert.Equal(t, 2, len(c.ExternalDatabases))

}
