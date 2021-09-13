package eacl

import (
	"go.uber.org/zap"
)

// Option represents Validator option.
type Option func(*cfg)

type cfg struct {
	logger *zap.Logger

	storage Source
}

func defaultCfg() *cfg {
	return &cfg{
		logger: zap.L(),
	}
}

// WithLogger configures the Validator to use logger v.
func WithLogger(v *zap.Logger) Option {
	return func(c *cfg) {
		c.logger = v
	}
}

// WithEACLSource configures the validator to use v as eACL source.
func WithEACLSource(v Source) Option {
	return func(c *cfg) {
		c.storage = v
	}
}
