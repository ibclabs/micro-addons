package rabbitmq

//
// All credit to Mondo
//

import (
	"errors"

	"github.com/pborman/uuid"
	"github.com/streadway/amqp"
)

type rmqChannel struct {
	uuid       string
	connection *amqp.Connection
	channel    *amqp.Channel
}

func newRabbitChannel(conn *amqp.Connection) (*rmqChannel, error) {
	rabbitCh := &rmqChannel{
		uuid:       uuid.NewRandom().String(),
		connection: conn,
	}
	if err := rabbitCh.Connect(); err != nil {
		return nil, err
	}
	return rabbitCh, nil

}

func (r *rmqChannel) Connect() error {
	var err error
	r.channel, err = r.connection.Channel()
	return err
}

func (r *rmqChannel) Close() error {
	if r.channel == nil {
		return errors.New("channel is nil")
	}
	return r.channel.Close()
}

func (r *rmqChannel) QoS(prefetchCount, prefetchSize int) error {
	if r.channel == nil {
		return errors.New("channel is nil")
	}
	return r.channel.Qos(prefetchCount, prefetchSize, false)
}

func (r *rmqChannel) Publish(exchange, key string, message amqp.Publishing) error {
	if r.channel == nil {
		return errors.New("channel is nil")
	}
	return r.channel.Publish(exchange, key, false, false, message)
}

func (r *rmqChannel) DeclareExchange(exchange string, durable bool) error {
	return r.channel.ExchangeDeclare(
		exchange, // name
		"topic",  // kind
		durable,    // durable
		false,    // autoDelete
		false,    // internal
		false,    // noWait
		nil,      // args
	)
}

func (r *rmqChannel) DeclareQueue(queue string, durable bool) error {
	_, err := r.channel.QueueDeclare(
		queue, // name
		durable, // durable
		!durable,  // autoDelete
		false, // exclusive
		false, // noWait
		nil,   // args
	)
	return err
}

func (r *rmqChannel) ConsumeQueue(queue string, autoAck bool) (<-chan amqp.Delivery, error) {
	return r.channel.Consume(
		queue,   // queue
		r.uuid,  // consumer
		autoAck, // autoAck
		false,   // exclusive
		false,   // nolocal
		false,   // nowait
		nil,     // args
	)
}

func (r *rmqChannel) BindQueue(queue, key, exchange string, args amqp.Table) error {
	return r.channel.QueueBind(
		queue,    // name
		key,      // key
		exchange, // exchange
		false,    // noWait
		args,     // args
	)
}
