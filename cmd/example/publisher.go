package main

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type MsgBroker struct {
	conn *amqp.Connection
	queue string
}

func newMsgBroker(conn *amqp.Connection, queus string) (*MsgBroker, error) {
	emitter := &MsgBroker{
		conn: conn,
		queue: queus,
	}

	err := emitter.setup()
	if err != nil {
		return nil, err
	}

	return emitter, nil
}

func (a *MsgBroker) setup() error {
	channel, err := a.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(
		a.queue,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return err
	}

	err = channel.Qos(
		1,
		0,
		false,
	)

	return err
}

func (a *MsgBroker) PublishLog(log any) error {
	wireData, err := json.Marshal(log)
	if err != nil {
		return err
	}

	channel, err := a.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	msg := amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		Body:         wireData,
		ContentType:  "application/json",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return channel.PublishWithContext(
		ctx,
		"",
		a.queue,
		false,
		false,
		msg,
	)
}
