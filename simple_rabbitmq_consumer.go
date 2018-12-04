package manager

import (
	"sync"

	"github.com/streadway/amqp"
)

type SimpleRabbitmqConsumer struct {
	config     *RabbitmqConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	queue      string
	bindingKey string
	tag        string
	handler    RabbitmqHandler
	done       chan error
	started    bool
}

func NewSimpleRabbitmqConsumer(config *RabbitmqConfig, queue, bindingKey, tag string, handler RabbitmqHandler) (*SimpleRabbitmqConsumer, error) {
	consumer := &SimpleRabbitmqConsumer{
		config:     config,
		connection: nil,
		channel:    nil,
		queue:      queue,
		bindingKey: bindingKey,
		tag:        tag,
		handler:    handler,
		done:       make(chan error),
	}

	return consumer, nil
}

func (consumer *SimpleRabbitmqConsumer) Start(wg *sync.WaitGroup) error {
	if wg == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	defer wg.Done()

	if consumer.started {
		return nil
	}

	var err error
	consumer.connection, err = consumer.config.Connect()
	if err != nil {
		err = log.Errorf("dial: %s", err).ToError()
		return err
	}

	defer func(err error) {
		if err != nil {
			if consumer.connection != nil {
				consumer.connection.Close()
			}
		}
	}(err)

	log.Infof("got connection, getting channel")
	consumer.channel, err = consumer.connection.Channel()
	if err != nil {
		err = log.Errorf("channel: %s", err).ToError()
		return err
	}

	log.Infof("got channel, declaring exchange (%s)", consumer.config.Exchange)
	if err = consumer.channel.ExchangeDeclare(
		consumer.config.Exchange,     // name of the exchange
		consumer.config.ExchangeType, // type
		true,  // durable
		false, // delete when complete
		false, // internal
		false, // noWait
		nil,   // arguments
	); err != nil {
		err = log.Errorf("exchange declare: %s", err).ToError()
		return err
	}

	log.Infof("declared exchange, declaring queue (%s)", consumer.queue)
	var queue amqp.Queue
	if queue, err = consumer.channel.QueueDeclare(
		consumer.queue, // name of the queue
		true,           // durable
		false,          // delete when usused
		false,          // exclusive
		false,          // noWait
		nil,            // arguments
	); err != nil {
		err = log.Errorf("queue declare: %s", err).ToError()
		return err
	}

	log.Infof("declared queue (%d messages, %d consumers), binding to exchange (bindingKey '%s')", queue.Messages, queue.Consumers, consumer.bindingKey)

	if err = consumer.channel.QueueBind(
		consumer.queue,           // name of the queue
		consumer.bindingKey,      // bindingKey
		consumer.config.Exchange, // sourceExchange
		false, // noWait
		nil,   // arguments
	); err != nil {
		err = log.Errorf("queue bind: %s", err).ToError()
		return err
	}

	log.Infof("queue bound to exchange, starting consume (consumer tag '%s')", consumer.tag)
	var deliveries <-chan amqp.Delivery
	if deliveries, err = consumer.channel.Consume(
		consumer.queue, // name
		consumer.tag,   // consumerTag,
		false,          // noAck
		false,          // exclusive
		false,          // noLocal
		false,          // noWait
		nil,            // arguments
	); err != nil {
		err = log.Errorf("queue consume: %s", err).ToError()
		return err
	}

	go consumer.handle(deliveries, consumer.done)

	consumer.started = true

	return nil
}

func (consumer *SimpleRabbitmqConsumer) Started() bool {
	return consumer.started
}

func (consumer *SimpleRabbitmqConsumer) Stop(wg *sync.WaitGroup) error {
	if wg == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	defer wg.Done()

	if !consumer.started {
		return nil
	}

	// will close() the deliveries channel
	if err := consumer.channel.Cancel(consumer.tag, true); err != nil {
		err = log.Errorf("consumer cancel failed: %s", err).ToError()
		return err
	}

	if err := consumer.connection.Close(); err != nil {
		err = log.Errorf("AMQP connection close error: %s", err).ToError()
		return err
	}

	log.Infof("AMQP shutdown OK")

	consumer.started = false

	// wait for handle() to exit
	return <-consumer.done

}

func (consumer *SimpleRabbitmqConsumer) handle(deliveries <-chan amqp.Delivery, done chan error) {
	for delivery := range deliveries {
		if err := consumer.handler(delivery); err != nil {
			delivery.Ack(false)
		} else {
			delivery.Ack(false)
		}
		log.Infof("got %dB delivery: [%v] %s", len(delivery.Body), delivery.DeliveryTag, delivery.Body)
	}

	log.Infof("handle: deliveries channel closed")
	done <- nil
}
