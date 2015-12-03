ToDD - Message Queue Design
====

This document outlines the design for using message queues in ToDD

# Message Queue Abstraction

Message queue implementations in ToDD offer two things:

- Control plane operations (setting up queues, exchanges, etc)
- Data plane operations (sending data into queue, listening for messages on queue)

Both of these features are implemented using an abstraction layer, such that it should not matter what specific queue implementation is used, ToDD daemons leverage this layer just the same.

(Golang interfaces are used here)

# Message Queue Requirements

Here are some things that ToDD needs the message queue for:

Alpha
- Agent Advertisements (direct to server)
- Fact Remediation (direct to agent)

Future
- App Remediation (direct to agent) ???

# RabbitMQ

It looks like you'll need to manually make bindings happen on teh queue level if you want to be specific. That means you'll have to integrate with RabbitMQ to create the exchanges and bindings you need. This will have to be extended to other message queues as well.
http://hg.rabbitmq.com/rabbitmq-management/raw-file/3646dee55e02/priv/www-api/help.html