package common

import (
	"context"
	"errors"
	"io"
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
	MaxBatch      int
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
	ctx    context.Context
	reader *BetReader
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig, ctx context.Context) *Client {

	client := &Client{
		config: config,
		ctx:    ctx,
		reader: NewBetReader(config.MaxBatch, config.ID),
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
	for !c.reader.Finished() {
		// Create the connection the server in every loop iteration. Send an
		if c.createClientSocket() != nil {
			return
		}

		c.ctx.(*signalCtx).SetConnection(&c.conn)

		bets, err := c.reader.ReadBets()
		if err != nil {
			log.Errorf("action: read_bets | result: fail| error: %v", err)
			c.conn.Close()
			return
		}

		if len(bets) == 0 {
			c.conn.Close()
			break loop
		}

		err = SendBets(c.conn, bets, c.config.ID)
		if err != nil {
			error_handler(err, "apuesta_envidad", c.ctx)
			close_conn_if_alive(c.ctx, err, &c.conn)
			return
		}

		answer, err := RecvAnswer(c.conn)
		if err != nil {
			error_handler(err, "respuesta_recibida", c.ctx)
			close_conn_if_alive(c.ctx, err, &c.conn)
			return
		}

		if answer == SUCCESS {
			log.Infof("action: apuesta_enviada | result: success | cantidad: %v", len(bets))
		} else {
			log.Infof("action: apuesta_enviada | result: fail | cantidad: %v", len(bets))
		}

		close_conn_if_alive(c.ctx, err, &c.conn)

		select {
		case <-c.ctx.Done():
			break loop
			// adds sleep
		case <-time.After(c.config.LoopPeriod):
			continue
		}

	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func close_conn_if_alive(ctx context.Context, err error, conn *net.Conn) {
	if errors.Is(err, net.ErrClosed) {
		return
	}
	if errors.Is(err, io.EOF) {
		log.Infof("action: server_closed_connection| result: success")
		return
	}
	select {
	case <-ctx.Done():
		return
	default:
		(*conn).Close()
	}
}

func error_handler(err error, message string, ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	default:
		if !errors.Is(err, net.ErrClosed) && !errors.Is(err, io.EOF) {
			log.Criticalf(
				"action: %s | result: fail | error: %v",
				message,
				err,
			)
			return

		}
	}
}
