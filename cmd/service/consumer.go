package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type msgBroker struct {
	conn  *amqp.Connection
	queue string
}

func (a *msgBroker) setup() error {
	channel, err := a.conn.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	_, err = channel.QueueDeclare(a.queue, true, false, false, false, nil)
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

func newMsgBroker(conn *amqp.Connection, queue string) (*msgBroker, error) {
	listener := &msgBroker{
		conn:  conn,
		queue: queue,
	}

	err := listener.setup()
	if err != nil {
		return nil, err
	}

	return listener, nil
}

// processLogs receives and process the logs published
// from different services.
func (svc *service) processLogs() {
	channel, err := svc.msgBroker.conn.Channel()
	if err != nil {
		svc.logger.PrintFatal(err, nil)
		return
	}
	defer channel.Close()

	msgs, err := channel.Consume(svc.msgBroker.queue, "", false, false, false, false, nil)
	if err != nil {
		svc.logger.PrintFatal(err, nil)
		return
	}

	var block chan struct{}

	go func() {
		for msg := range msgs {
			var log model.Log

			err := json.Unmarshal(msg.Body, &log)
			if err != nil {
				svc.logger.PrintError(err, map[string]string{
					"type":  "failed to unmarshal json",
					"error": fmt.Sprintf("%+v", string(msg.Body)),
				})
				continue
			}

			v := utils.NewValidator()
			if utils.ValidateLog(v, &log); !v.Valid() {
				svc.logger.PrintError(errors.New("failed to validate log"), map[string]string{
					"errors": fmt.Sprintf("%+v", v.Errors),
					"log":    fmt.Sprintf("%+v", log),
				})
				continue
			}

			id, err := svc.logs.AddLog(&log)
			if err != nil {
				svc.logger.PrintError(err, map[string]string{
					"type": "failed to write log",
					"log":  fmt.Sprintf("%+v", log),
				})
				continue
			}

			svc.logger.PrintInfo("log added to data store", map[string]string{
				"resource_id": fmt.Sprintf("%+v", id),
			})

			msg.Ack(false)
		}
	}()

	svc.logger.PrintInfo("Waiting for log messages", nil)
	<-block
}
