package rabbitmq

import (
	"context"

	"github.com/micro/go-micro/broker"
)

type durableQueueKey struct{}
type headersKey struct{}
type prefetchCountKey struct{}
type exchangeNameKey struct{}
type durableExchangeKey struct{}

// DurableQueue creates a durable queue when subscribing.
func DurableQueue() broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, durableQueueKey{}, true)
	}
}

// Headers adds headers used by the headers exchange
func Headers(h map[string]interface{}) broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, headersKey{}, h)
	}
}

// PrefetchCount creates an option for rmq QoS.
func PrefetchCount(c int) broker.SubscribeOption {
	return func(o *broker.SubscribeOptions) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, prefetchCountKey{}, c)
	}
}

// Exchange is an option to set the Exchange
func Exchange(e string) broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, exchangeNameKey{}, e)
	}
}

// DurableExchange creates an option for durable exchange when subscribing.
func DurableExchange() broker.Option {
	return func(o *broker.Options) {
		if o.Context == nil {
			o.Context = context.Background()
		}
		o.Context = context.WithValue(o.Context, durableExchangeKey{}, true)
	}
}
