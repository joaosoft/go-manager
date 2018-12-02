package manager

import (
	"time"

	"sync"

	"github.com/streadway/amqp"
)

type SimpleRabbitmqProducer struct {
	config     *RabbitmqConfig
	connection *amqp.Connection
	channel    *amqp.Channel
	tag        string
	started    bool
}

func NewSimpleRabbitmqProducer(config *RabbitmqConfig) (*SimpleRabbitmqProducer, error) {
	return &SimpleRabbitmqProducer{
		config: config,
	}, nil
}

func (producer *SimpleRabbitmqProducer) Start(wg *sync.WaitGroup) error {
	if wg == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	defer wg.Done()

	if producer.started {
		return nil
	}

	producer.started = true
	var err error
	producer.connection, err = producer.config.Connect()
	if err != nil {
		err = log.Errorf("dial: %s", err).ToError()
		return err
	}

	defer func(err error) {
		if err != nil {
			if producer.connection != nil {
				producer.connection.Close()
			}
		}
	}(err)

	log.Infof("got connection, getting channel")
	if producer.channel, err = producer.connection.Channel(); err != nil {
		err = log.Errorf("channel: %s", err).ToError()
		return err
	}

	log.Infof("got channel, declaring %q exchange (%s)", producer.config.ExchangeType, producer.config.Exchange)
	if err = producer.channel.ExchangeDeclare(
		producer.config.Exchange,     // name
		producer.config.ExchangeType, // type
		true,  // durable
		false, // auto-deleted
		false, // internal
		false, // noWait
		nil,   // arguments
	); err != nil {
		err = log.Errorf("exchange declare: %s", err).ToError()
		return err
	}

	return nil
}

func (producer *SimpleRabbitmqProducer) Started() bool {
	producer.started = false
	return producer.started
}

func (producer *SimpleRabbitmqProducer) Stop(wg *sync.WaitGroup) error {
	if wg == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	defer wg.Done()

	if !producer.started {
		return nil
	}

	producer.started = false
	// will close() the deliveries channel
	if err := producer.channel.Cancel(producer.tag, true); err != nil {
		err = log.Errorf("consumer cancel failed: %s", err).ToError()
		return err
	}

	if err := producer.connection.Close(); err != nil {
		err = log.Errorf("AMQP connection close error: %s", err).ToError()
		return err
	}

	log.Infof("AMQP shutdown OK")

	return nil
}

func (producer *SimpleRabbitmqProducer) Publish(routingKey string, body []byte, reliable bool) error {
	msg := amqp.Publishing{
		DeliveryMode:    amqp.Persistent,
		Timestamp:       time.Now(),
		Headers:         amqp.Table{},
		ContentType:     "text/plain",
		ContentEncoding: "",
		Body:            body,
		Priority:        0, // 0-9
	}

	log.Infof("declared exchange, publishing %dB body (%s)", len(body), body)
	if err := producer.channel.Publish(
		producer.config.Exchange, // publish to an exchange
		routingKey,               // routing to 0 or more queues
		false,                    // mandatory
		false,                    // immediate
		msg,
	); err != nil {
		err = log.Errorf("exchange publish: %s", err).ToError()
		return err
	}

	return nil
}
