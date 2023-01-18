package main

import (
	"flag"
	"os"
	"time"

	"github.com/IkehAkinyemi/logaudit/internal/jsonlog"
	"github.com/IkehAkinyemi/logaudit/internal/repository/model"
	"github.com/IkehAkinyemi/logaudit/internal/utils"
	amqp "github.com/rabbitmq/amqp091-go"
)

type log struct {
	Timestamp time.Time      `json:"created_at"`
	Action    string         `json:"action"`
	Actor     model.Actor    `json:"actor"`
	Entity    model.Entity   `json:"entity"`
	Context   model.Context  `json:"context"`
	Extension map[string]any `json:"extension,omitempty"`
}

func main() {
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Parse flag argument
	var cfgFile string
	flag.StringVar(&cfgFile, "cfg-file", "", "Directory path to configuration file")
	flag.Parse()

	config, err := utils.ParseConfig()
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}

	// connect message broker
	conn, err := amqp.Dial(config.AMQP_CONN_URI)
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}
	logger.PrintInfo("connection to message broker established", nil)
	defer conn.Close()

	msgBroker, err := newMsgBroker(conn, "logs")
	if err != nil {
		logger.PrintFatal(err, nil)
		return
	}

	msg := &log{
		Timestamp: time.Now(),
		Action:    "updated",
		Actor: model.Actor{
			Type:      "user",
			ID:        "12300",
			Extension: map[string]any{"userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/83.0.4103.116 Safari/537.36"},
		},
		Entity: model.Entity{
			Type:      "inventory",
			Extension: map[string]any{"item_id": "f66020564728"},
		},
		Context: model.Context{
			IPAddr:    "192.168.1.1",
			Location:  "New York, NY",
			Extension: map[string]any{"inventory_section": "electronics"},
		},
		Extension: map[string]any{"notes": "Inventory successfully updated"},
	}

	if err := msgBroker.PublishLog(msg); err != nil {
		logger.PrintError(err, nil)
	}

	logger.PrintInfo("log published to message broker", nil)
}
