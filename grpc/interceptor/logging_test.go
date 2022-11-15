package interceptor

import (
	"testing"

	"github.com/zhufuyi/pkg/logger"

	"github.com/stretchr/testify/assert"
)

func TestStreamClientLog(t *testing.T) {
	interceptor := StreamClientLog(logger.Get())
	assert.NotNil(t, interceptor)
}

func TestStreamServerCtxTags(t *testing.T) {
	interceptor := StreamServerCtxTags()
	assert.NotNil(t, interceptor)
}

func TestStreamServerLog(t *testing.T) {
	interceptor := StreamServerLog(nil,
		WithLogFields(map[string]interface{}{"foo": "bar"}),
		WithLogIgnoreMethods("/ping"),
	)
	assert.NotNil(t, interceptor)
}

func TestUnaryClientLog(t *testing.T) {
	interceptor := UnaryClientLog(logger.Get())
	assert.NotNil(t, interceptor)
}

func TestUnaryServerCtxTags(t *testing.T) {
	interceptor := UnaryServerCtxTags()
	assert.NotNil(t, interceptor)
}

func TestUnaryServerLog(t *testing.T) {
	interceptor := UnaryServerLog(nil,
		WithLogFields(map[string]interface{}{"foo": "bar"}),
		WithLogIgnoreMethods("/ping"),
	)
	assert.NotNil(t, interceptor)
}
