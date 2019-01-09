package rabbitmq

//
// All credit to Mondo
//

import (
	"crypto/tls"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

var (
	defaultExchange  = "micro"
	defaultRabbitURL = "amqp://guest:guest@127.0.0.1:5672"

	dial    = amqp.Dial
	dialTLS = amqp.DialTLS
)

type rmqConnection struct {
	Connection      *amqp.Connection
	Channel         *rmqChannel
	exchange        rmqExchange
	url             string

	sync.Mutex
	connected bool
	close     chan bool
}

type rmqExchange struct {
	Name string
	Durable bool
}

func newRabbitConnection(exchange rmqExchange, urls []string) *rmqConnection {
	var url string

	if len(urls) > 0 && regexp.MustCompile("^amqp(s)?://.*").MatchString(urls[0]) {
		url = urls[0]
	} else {
		url = defaultRabbitURL
	}

	if len(exchange.Name) == 0 {
		exchange.Name = defaultExchange
	}

	return &rmqConnection{
		exchange: exchange,
		url:      url,
		close:    make(chan bool),
	}
}

func (r *rmqConnection) connect(secure bool, config *tls.Config) error {
	// try connect
	if err := r.tryConnect(secure, config); err != nil {
		return err
	}

	// connected
	r.Lock()
	r.connected = true
	r.Unlock()

	// create reconnect loop
	go r.reconnect(secure, config)
	return nil
}

func (r *rmqConnection) reconnect(secure bool, config *tls.Config) {
	// skip first connect
	var connect bool

	for {
		if connect {
			// try reconnect
			if err := r.tryConnect(secure, config); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			// connected
			r.Lock()
			r.connected = true
			r.Unlock()
		}

		connect = true
		notifyClose := make(chan *amqp.Error)
		r.Connection.NotifyClose(notifyClose)

		// block until closed
		select {
		case <-notifyClose:
			r.Lock()
			r.connected = false
			r.Unlock()
		case <-r.close:
			return
		}
	}
}

func (r *rmqConnection) Connect(secure bool, config *tls.Config) error {
	r.Lock()

	// already connected
	if r.connected {
		r.Unlock()
		return nil
	}

	// check it was closed
	select {
	case <-r.close:
		r.close = make(chan bool)
	default:
		// no op
		// new conn
	}

	r.Unlock()

	return r.connect(secure, config)
}

func (r *rmqConnection) Close() error {
	r.Lock()
	defer r.Unlock()

	select {
	case <-r.close:
		return nil
	default:
		close(r.close)
		r.connected = false
	}

	return r.Connection.Close()
}

func (r *rmqConnection) tryConnect(secure bool, config *tls.Config) error {
	var err error

	if secure || config != nil || strings.HasPrefix(r.url, "amqps://") {
		if config == nil {
			config = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		url := strings.Replace(r.url, "amqp://", "amqps://", 1)
		r.Connection, err = dialTLS(url, config)
	} else {
		r.Connection, err = dial(r.url)
	}

	if err != nil {
		return err
	}

	if r.Channel, err = newRabbitChannel(r.Connection); err != nil {
		return err
	}

	return r.Channel.DeclareExchange(r.exchange.Name, r.exchange.Durable)
}

func (r *rmqConnection) Consume(queue, key string, headers amqp.Table, autoAck, durableQueue bool, prefetchCount int) (*rmqChannel, <-chan amqp.Delivery, error) {
	consChannel, err := newRabbitChannel(r.Connection)
	if err != nil {
		return nil, nil, err
	}

	if prefetchCount > 0 {
		err = consChannel.QoS(prefetchCount, 0)
		if err != nil {
			return nil, nil, err
		}
	}

	err = consChannel.DeclareQueue(queue, durableQueue)
	if err != nil {
		return nil, nil, err
	}

	deliveries, err := consChannel.ConsumeQueue(queue, autoAck)
	if err != nil {
		return nil, nil, err
	}

	err = consChannel.BindQueue(queue, key, r.exchange.Name, headers)
	if err != nil {
		return nil, nil, err
	}

	return consChannel, deliveries, nil
}

func (r *rmqConnection) Publish(exchange, key string, msg amqp.Publishing) error {
	return r.Channel.Publish(exchange, key, msg)
}
