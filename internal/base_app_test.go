package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseApp(t *testing.T) {

	d := NewBaseApp()

	assert.NotNil(t, d.logger)
}

func TestStartBaseApp(t *testing.T) {
	d := NewBaseApp()

	d.Bootstrap()

	assert.NotNil(t, d.logger)
	assert.NotNil(t, d.onAfterApplicationBootstrapped)
	assert.NotNil(t, d.data)
}
