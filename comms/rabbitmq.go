/*
   ToDD commsPackage implementation for RabbitMQ

   Copyright 2015 - Matt Oswalt
*/

package comms

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	log "github.com/mierdin/todd/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/mierdin/todd/Godeps/_workspace/src/github.com/streadway/amqp"
	"github.com/mierdin/todd/agent/defs"
	"github.com/mierdin/todd/common"
	"github.com/mierdin/todd/config"
	"github.com/mierdin/todd/db"
)

type RabbitMQComms struct {
	config config.Config
}

// AdvertiseAgent will place an agent advertisement message on the message queue
func (rmq RabbitMQComms) AdvertiseAgent(me defs.AgentAdvert) {

	// TODO: Need to change something here so that the messages get
	// dropped if there's no one listening. As it stands, these messages
	// persist until the server is started, then they're all delivered at once

	queue_url := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		rmq.config.AMQP.User,
		rmq.config.AMQP.Password,
		rmq.config.AMQP.Host,
		rmq.config.AMQP.Port,
	)

	conn, err := amqp.Dial(queue_url)
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"test_exchange", // name
		"direct",        // kind
		false,           // durable
		false,           // delete when unused
		false,           // internal
		false,           // no-wait
		nil,             // args
	)
	common.FailOnError(err, "Failed to declare an exchange")

	q, err := ch.QueueDeclare(
		"agentadvert", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	common.FailOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		"agentadvert",   // name
		"agentadvert",   // routing key
		"test_exchange", // exchange
		false,           // no-wait
		nil,             // args
	)
	common.FailOnError(err, "Failed to bind exchange to queue")

	// Marshal agent struct to JSON
	json_data, err := json.Marshal(me)
	common.FailOnError(err, "Failed to marshal object data")

	err = ch.Publish(
		"test_exchange", // exchange
		q.Name,          // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(json_data),
		})
	common.FailOnError(err, "Failed to publish a message")

	log.Infof("AGENTADV -- %s", time.Now().UTC())
}

// ListenForAgent will listen on the message queue for new agent advertisements.
// It is meant to be run as a goroutine
func (rmq RabbitMQComms) ListenForAgent(fact_collectors map[string]string) {

	// TODO(mierdin): does func param need to be a pointer?

	queue_url := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		rmq.config.AMQP.User,
		rmq.config.AMQP.Password,
		rmq.config.AMQP.Host,
		rmq.config.AMQP.Port,
	)

	conn, err := amqp.Dial(queue_url)
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"agentadvert", // name
		false,         // durable
		false,         // delete when usused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	common.FailOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name,        // queue
		"agentadvert", // consumer
		true,          // auto-ack
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	common.FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Debugf("MESSAGE RECEIVED: %s", d.Body)

			var agent defs.AgentAdvert
			err = json.Unmarshal(d.Body, &agent)
			// TODO(mierdin): Need to handle this error

			// remediateList is a slice that will contain any URLs that need to be sent to an
			// agent as a response to an incorrect or incomplete list of fact collectors
			var remediateList []string

			// fact_collectors is the plug map from the SERVER's perspective
			// agent.FactCollectors is the collector map from the AGENT's perspective
			for name, _ := range fact_collectors {

				// See if the hashes match
				if agent.FactCollectors[name] != fact_collectors[name] {

					// hashes do not match, append this as a URL to the remediate list
					collector_url := fmt.Sprintf("http://%s:%s/%s", rmq.config.Facts.IP, rmq.config.Facts.Port, name)
					remediateList = append(remediateList, collector_url)
				}
			}

			// Fact remediation list is empty - perform database insertion
			if len(remediateList) == 0 {

				// TODO(mierdin): better config
				cfg := config.GetConfig("/etc/server_config.cfg")
				var tdb = db.NewToddDB(cfg)
				tdb.DatabasePackage.SetAgent(agent)

			} else { // List not empty - send remediation notice
				rmq.SendFactRemediationNotice(agent.Uuid, remediateList)
			}

		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

// SendFactRemediationNotice will send a direct message to an agent that returned an incorrect list of fact collectors
func (rmq RabbitMQComms) SendFactRemediationNotice(uuid string, remediations []string) {

	// TODO(mierdin): Need to change something here so that the messages get
	// dropped if there's no one listening. As it stands, these messages
	// persist until the server is started, then they're all delivered at once

	queue_url := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		rmq.config.AMQP.User,
		rmq.config.AMQP.Password,
		rmq.config.AMQP.Host,
		rmq.config.AMQP.Port,
	)

	conn, err := amqp.Dial(queue_url)
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.ExchangeDeclare(
		"test_exchange", // name
		"direct",        // kind
		false,           // durable
		false,           // delete when unused
		false,           // internal
		false,           // no-wait
		nil,             // args
	)
	common.FailOnError(err, "Failed to declare an exchange")

	_, err = ch.QueueDeclare(
		uuid,  // name
		false, // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	common.FailOnError(err, "Failed to declare a queue")

	err = ch.QueueBind(
		uuid,            // name
		uuid,            // routing key
		"test_exchange", // exchange
		false,           // no-wait
		nil,             // args
	)
	common.FailOnError(err, "Failed to bind exchange to queue")

	// Marshal remediation slice to JSON
	remediation_msg := make(map[string][]string)
	remediation_msg["FactRemediation"] = remediations
	json_data, err := json.Marshal(remediation_msg)
	common.FailOnError(err, "Failed to marshal object data")

	err = ch.Publish(
		"test_exchange", // exchange
		uuid,            // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(json_data),
		})
	common.FailOnError(err, "Failed to publish a message")

	log.Infof("SENT FACT REMEDIATION TO -- %s", uuid)
}

func (rmq RabbitMQComms) ListenForFactRemediationNotice(uuid string) {

	queue_url := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		rmq.config.AMQP.User,
		rmq.config.AMQP.Password,
		rmq.config.AMQP.Host,
		rmq.config.AMQP.Port,
	)

	conn, err := amqp.Dial(queue_url)
	common.FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	common.FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	_, err = ch.QueueDeclare(
		uuid,  // name
		false, // durable
		false, // delete when usused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	common.FailOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		uuid,  // queue
		uuid,  // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	common.FailOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {

			log.Debugf("REMEDIATION RECEIVED: %s", d.Body)

			remediation_msg := make(map[string][]string)

			err = json.Unmarshal(d.Body, &remediation_msg)
			// TODO(mierdin): Need to handle this error

			for x := range remediation_msg["FactRemediation"] {
				downloadCollector(remediation_msg["FactRemediation"][x], rmq.config.Facts.CollectorDir)
			}
		}
	}()

	log.Infof(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func downloadCollector(url string, destDir string) {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	fileName = fmt.Sprintf("%s/%s", destDir, fileName)
	fmt.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist
	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	err = os.Chmod(fileName, 0744)
	common.FailOnError(err, "Problem setting execute permission on downloaded script")

	fmt.Println(n, "bytes downloaded.")
}
