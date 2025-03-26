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

func (c *Client) Start() {
	defer c.reader.Close()
	for {
		if c.createClientSocket() != nil {
			return
		}
		c.ctx.(*signalCtx).SetConnection(&c.conn)

		if !c.reader.Finished() {
			err := c.sendBets()
			if err != nil {
				return
			}
			select {
			case <-c.ctx.Done():
				log.Info("action: shutting down| result: success")
				return
			default:
				continue
			}

		} else {
			err := c.awaitResults()
			if err != nil {
				return
			}
			break
		}

	}
	// This sleep is explicitly added for testing. 
	time.Sleep(c.config.LoopPeriod)
	log.Info("action: client_finished | result: success")
}

func (c *Client) sendBets() error {
	defer c.conn.Close()
	bets, err := c.reader.ReadBets()
	if err != nil {
		error_handler(err, "read_bets", c.ctx)
		return err
	}

	if len(bets) == 0 {
		return nil
	}

	err = SendBets(c.conn, bets, c.config.ID)
	if err != nil {
		error_handler(err, "sending_bets", c.ctx)
		return err
	}

	answer, err := RecvAnswer(c.conn)
	if err != nil {
		error_handler(err, "respuesta_recibida", c.ctx)
		return err
	}

	if answer == SUCCESS {
		log.Infof("action: apuesta_enviada | result: success | cantidad: %v", len(bets))
	} else {
		log.Infof("action: apuesta_enviada | result: fail | cantidad: %v", len(bets))
	}
	return nil
}

func (c *Client) awaitResults() error {
	defer c.conn.Close()

	err := SendEndMessage(c.conn, c.config.ID)
	if err != nil {
		error_handler(err, "Send End", c.ctx)
		return err
	}

	winners, err := RecvResults(c.conn)
	if err != nil {
		error_handler(err, "awaiting results", c.ctx)
		return err
	} else {
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %v", len(winners))
	}

	err = sendFinish(c.conn)
	if err != nil {
		error_handler(err, "finishing", c.ctx)
		return err
	}
	return nil
}

func error_handler(err error, message string, ctx context.Context) {
	if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
		log.Infof("action: %s | result: fail | reason: sever_closed", message)
		log.Infof("action: shutting_down | result: success")
		return
	}
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
		}
	}
}
