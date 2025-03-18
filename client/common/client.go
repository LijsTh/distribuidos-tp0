package common

import (
	"context"
	"net"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	ctx    context.Context
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, ctx context.Context) *Client {
	client := &Client{
		config: config,
		ctx:    ctx,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop() {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
loop:
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		if c.createClientSocket() != nil {
			return
		}

		bet, err := NewBetFromEnv()
		if err != nil {
			log.Errorf("action: read_bets | result: fail| error: %v", err)
			c.conn.Close()
			return
		}

		err = SendBet(c.conn, bet)
		if err != nil {
			log.Errorf("action: apuesta_enviada | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.conn.Close()
			return
		}

		answer, err := RecvAnswer(c.conn)
		if err != nil {
			log.Errorf("action: respuesta_recibida | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.conn.Close()
			return
		}

		if answer == SUCESS {
			log.Infof("action: apuesta_enviada | result: success | dni: %v | number: %v", bet.document, bet.number)
		} else {
			log.Infof("action: apuesta_enviada | result: fail | dni: %v | number: %v", bet.document, bet.number)
		}

		c.conn.Close()

		select {
		case <-c.ctx.Done():
			break loop
		case <-time.After(c.config.LoopPeriod): // DEFAULT later
			continue
		}

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}
