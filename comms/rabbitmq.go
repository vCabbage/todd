/*
    ToDD commsPackage implementation for RabbitMQ

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package comms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/config"
)

const (
	connectRetry = 3

	// Queue Names
	queueAgentAdvertise = "agentadvert"
	queueAgentResponses = "agentresponses"
)

func init() {
	register("rabbitmq", newRabbitMQComms)
}

type rabbitMQComms struct {
	config *config.Config
	conn   *amqp.Connection

	exchange string
}

// newRabbitMQComms is a factory function that produces a new instance of rabbitMQComms with the configuration
// loaded and ready to be used.
func newRabbitMQComms(cfg *config.Config) (Comms, error) {
	rmq := &rabbitMQComms{
		config:   cfg,
		exchange: "test_exchange",
	}

	if err := rmq.connect(); err != nil {
		return nil, err
	}

	if err := rmq.init(); err != nil {
		return nil, err
	}

	return rmq, nil
}

// connectRabbitMQ wraps the amqp.Dial function in order to provide connection retry functionality
func (rmq *rabbitMQComms) connect() error {
	url := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		rmq.config.Comms.User,
		rmq.config.Comms.Password,
		rmq.config.Comms.Host,
		rmq.config.Comms.Port,
	)

	var err error
	for retries := 0; retries < connectRetry; retries++ {
		rmq.conn, err = amqp.Dial(url)
		if err == nil {
			break
		}

		log.Warnf("Failure connecting to RabbitMQ - retry #%d", retries)
		time.Sleep(1 * time.Second)
	}
	return err
}

func (rmq *rabbitMQComms) init() error {
	ch, err := rmq.conn.Channel()
	if err != nil {
		log.Error("Failed to open a channel")
		log.Debug(err)
		return err
	}
	defer ch.Close()

	// Create Exchange
	return ch.ExchangeDeclare(
		rmq.exchange, // name
		"direct",     // kind
		false,        // durable
		false,        // delete when unused
		false,        // internal
		false,        // no-wait
		nil,          // args
	)
}

// ListenForAgent will listen on the message queue for new agent advertisements.
// It is meant to be run as a goroutine
func (rmq *rabbitMQComms) ListenForAgent(ctx context.Context) (chan []byte, error) {
	return rmq.listenQueue(queueAgentAdvertise, ctx)
}

// ListenForResponses listens for responses from an agent
func (rmq *rabbitMQComms) ListenForResponses(ctx context.Context) (chan []byte, error) {
	return rmq.listenQueue(queueAgentResponses, ctx)
}

// SendTask will send a task object onto the specified queue ("to").
//
// This could be an agent UUID, or a group name. Agents that have been
// added to a group.
func (rmq *rabbitMQComms) SendTask(to string, task tasks.Task) error {
	return rmq.sendJSON(to, task)
}

// AdvertiseAgent will place an agent advertisement message on the message queue
func (rmq *rabbitMQComms) AdvertiseAgent(me defs.AgentAdvert) error {
	err := rmq.sendJSON(queueAgentAdvertise, me)
	return err
}

// ListenForTasks is a method that recieves task notices from the server
func (rmq *rabbitMQComms) ListenForTasks(from string, ctx context.Context) (chan []byte, error) {
	return rmq.listenQueue(from, ctx)
}

// sendResponse will send a response object onto the
// statically-defined queue for receiving such messages.
func (rmq *rabbitMQComms) SendResponse(resp interface{}) error {
	err := rmq.sendJSON(queueAgentResponses, resp)
	return errors.Wrap(err, "marshaling to JSON")
}

func (rmq *rabbitMQComms) listenQueue(queue string, ctx context.Context) (chan []byte, error) {
	ch, err := rmq.channel(queue)
	if err != nil {
		return nil, errors.Wrap(err, "opening channel")
	}

	msgs, err := ch.consume()
	if err != nil {
		return nil, errors.Wrap(err, "registering consumer")
	}

	responses := make(chan []byte)
	go func() {
		defer ch.close()
		for {
			select {
			case msg := <-msgs:
				responses <- msg.Body
			case <-ctx.Done():
				close(responses)
				return
			}
		}
	}()

	return responses, nil
}

func (rmq *rabbitMQComms) sendJSON(chanName string, v interface{}) error {
	ch, err := rmq.channel(chanName)
	if err != nil {
		return err
	}
	defer ch.close()

	jsonData, err := json.Marshal(v)
	if err != nil {
		return errors.Wrap(err, "marshaling to JSON")
	}

	err = ch.publish(jsonData)

	log.Debugf("Sent to %s: %+v", chanName, v)

	return errors.Wrap(err, "publishing to RabbitMQ")
}

func (rmq *rabbitMQComms) channel(name string) (*channel, error) {
	rmqCh, err := rmq.conn.Channel()
	if err != nil {
		return nil, errors.Wrap(err, "opening channel")
	}
	ch := &channel{ch: rmqCh, name: name, exchange: rmq.exchange}

	q, err := rmqCh.QueueDeclare(
		name,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		rmqCh.Close()
		return nil, errors.Wrap(err, "declaring queue")
	}

	err = rmqCh.QueueBind(
		q.Name,       // name
		name,         // routing key
		rmq.exchange, // exchange
		false,        // no-wait
		nil,          // args
	)
	if err != nil {
		rmqCh.Close()
		return nil, errors.Wrap(err, "binding queue")
	}

	return ch, nil
}

type channel struct {
	ch       *amqp.Channel
	name     string
	exchange string
}

func (c *channel) publish(message []byte) error {
	return c.ch.Publish(
		c.exchange, // exchange
		c.name,     // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Expiration:  "5000", // expiration in milliseconds (we don't want these messages to pile up if the server isn't running)
			Body:        message,
		},
	)
}

func (c *channel) consume() (<-chan amqp.Delivery, error) {
	return c.ch.Consume(
		c.name, // queue
		c.name, // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
}

func (c *channel) close() error {
	return c.ch.Close()
}
