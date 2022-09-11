package grpccli

import (
	"context"
	"errors"

	"projectExample/pkg/discovery"
	"projectExample/pkg/grpc/middleware"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/zhufuyi/pkg/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Dial 安全连接
func Dial(ctx context.Context, endpoint string, opts ...Option) (*grpc.ClientConn, error) {
	isSecure := true
	return dial(ctx, endpoint, isSecure, opts...)
}

// DialInsecure 不安全连接
func DialInsecure(ctx context.Context, endpoint string, opts ...Option) (*grpc.ClientConn, error) {
	isSecure := false
	return dial(ctx, endpoint, isSecure, opts...)
}

func dial(ctx context.Context, endpoint string, isSecure bool, opts ...Option) (*grpc.ClientConn, error) {
	o := defaultOptions()
	o.apply(opts...)

	var unaryClientInterceptors []grpc.UnaryClientInterceptor
	var streamClientInterceptors []grpc.StreamClientInterceptor

	var clientOptions []grpc.DialOption

	// 第一个clientOptions是服务发现
	if o.discovery != nil {
		clientOptions = append(clientOptions, grpc.WithResolvers(
			discovery.NewBuilder(
				o.discovery,
				discovery.WithInsecure(!isSecure),
			)))
	}

	// 是否安全连接
	if isSecure {
		if o.credentials == nil {
			return nil, errors.New("unset tls credentials")
		}
		clientOptions = append(clientOptions, grpc.WithTransportCredentials(o.credentials))
	} else {
		clientOptions = append(clientOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 日志
	if o.enableLog {
		unaryClientInterceptors = append(unaryClientInterceptors, middleware.UnaryClientLog(logger.Get()))
	}

	// 指标 metrics
	if o.enableMetrics {
		unaryClientInterceptors = append(unaryClientInterceptors, middleware.UnaryClientMetrics())
	}

	// 负载均衡器 load balance
	if o.enableLoadBalance {
		clientOptions = append(clientOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingConfig": [{"round_robin":{}}]}`))
	}

	// 熔断器 hystrix
	if o.enableHystrix {
		unaryClientInterceptors = append(unaryClientInterceptors, middleware.UnaryClientHystrix(o.hystrixName))
	}

	// 重试 retry
	if o.enableRetry {
		unaryClientInterceptors = append(unaryClientInterceptors, middleware.UnaryClientRetry())
	}

	unaryClientInterceptors = append(unaryClientInterceptors, o.unaryInterceptors...)
	streamClientInterceptors = append(streamClientInterceptors, o.streamInterceptors...)

	o.dialOptions = append(o.dialOptions,
		grpc.WithUnaryInterceptor(
			grpc_middleware.ChainUnaryClient(unaryClientInterceptors...),
		))
	o.dialOptions = append(o.dialOptions,
		grpc.WithStreamInterceptor(
			grpc_middleware.ChainStreamClient(streamClientInterceptors...),
		))

	clientOptions = append(clientOptions, o.dialOptions...)

	return grpc.DialContext(ctx, endpoint, clientOptions...)
}
