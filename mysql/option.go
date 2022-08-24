package mysql

import (
	"time"
)

// Option set the mysql options.
type Option func(*options)

type options struct {
	isLog         bool
	slowThreshold time.Duration

	maxIdleConns    int
	maxOpenConns    int
	connMaxLifetime time.Duration

	disableForeignKey bool
	enableTrace       bool
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// 默认设置
func defaultOptions() *options {
	return &options{
		isLog:         false,            // 是否输出日志，默认关闭
		slowThreshold: time.Duration(0), // 如果大于0，只打印时间大于阈值的日志，优先级比isLog高

		maxIdleConns:    3,               // 空闲连接数
		maxOpenConns:    30,              // 最大连接数
		connMaxLifetime: 5 * time.Minute, // 5分钟后断开多余的空闲连接

		disableForeignKey: true,  // 禁止使用外键，生产环境建议设置为true，默认开启
		enableTrace:       false, // 是否开启链路跟踪，默认关闭
	}
}

// WithLog set log sql
func WithLog() Option {
	return func(o *options) {
		o.isLog = true
	}
}

// WithSlowThreshold Set sql values greater than the threshold
func WithSlowThreshold(d time.Duration) Option {
	return func(o *options) {
		o.slowThreshold = d
	}
}

// WithMaxIdleConns set max idle conns
func WithMaxIdleConns(size int) Option {
	return func(o *options) {
		o.maxIdleConns = size
	}
}

// WithMaxOpenConns set max open conns
func WithMaxOpenConns(size int) Option {
	return func(o *options) {
		o.maxOpenConns = size
	}
}

// WithConnMaxLifetime set conn max liftime
func WithConnMaxLifetime(t time.Duration) Option {
	return func(o *options) {
		o.connMaxLifetime = t
	}
}

// WithEnableForeignKey use foreign keys
func WithEnableForeignKey() Option {
	return func(o *options) {
		o.disableForeignKey = false
	}
}

// WithEnableTrace use trace
func WithEnableTrace() Option {
	return func(o *options) {
		o.enableTrace = true
	}
}
