package service_test

import (
	"testing"

	service "goodin"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	instance := service.New(service.BaseConfig{
		DisableBanner: true,
	})

	assert.NotNil(t, instance.App)
	assert.NotNil(t, instance.RootCmd)
	assert.True(t, instance.TemplateConfig.DisableBanner)
}
