package common

import (
	"encoding/binary"
	"errors"
	"net"
	"strconv"
)

const AGENCY_SIZE = 1
const STR_SIZE = 1
const MAX_STR_SIZE = 255
const DOCUMENT_SIZE = 4
const BIRTHDATE_SIZE = 10
const NUMBER_SIZE = 2
const ANSWER_SIZE = 1
const BATCH_SIZE = 2

// result constants
const SUCCESS = 0
const FAIL = 1

// batch max
const BATCHMAX = 8000 //8kb

func RecvAll(conn net.Conn, size int) ([]byte, error) {
	/// Read all the bytes from the connection to avoid partial reads
	buf := make([]byte, size)
	total := 0

	for total < size {
		n, err := conn.Read(buf[total:])
		if err != nil || n == 0 {
			return nil, err
		}
		total += n
	}
	return buf, nil
}

func send_all(conn net.Conn, message []byte) error {
	/// Send all the bytes to the connection to avoid partial writes
	written := 0
	for written < len(message) {
		n, err := conn.Write(message)
		if err != nil || n == 0 {
			log.Criticalf(
				"action: send_all | result: fail| total read: %v |  error: %v",
				err,
			)
			return err
		}
		written += n
	}
	return nil
}

func serializeUnknownString(message string, buf []byte) []byte {
	if len(message) > MAX_STR_SIZE {
		log.Criticalf(
			"action: serialize_uknown_string | result: fail | error: string too long",
		)
		return nil
	}
	buf = append(buf, byte(len(message)))
	buf = append(buf, []byte(message)...)
	return buf
}

func encodeBet(bet *Bet) ([]byte, error) {
	msg := make([]byte, 0)

	// firstName
	msg = serializeUnknownString(bet.firstName, msg)
	if msg == nil {
		return nil, errors.New("error serializing firstName")
	}

	// lastName
	msg = serializeUnknownString(bet.lastName, msg)
	if msg == nil {
		return nil, errors.New("error serializing lastName")
	}

	// document
	docBytes := make([]byte, DOCUMENT_SIZE)
	binary.BigEndian.PutUint32(docBytes, bet.document)
	msg = append(msg, docBytes...)

	// birthDate
	msg = append(msg, []byte(bet.birthDate)...) // SIZE 10

	// number
	numBytes := make([]byte, NUMBER_SIZE)
	binary.BigEndian.PutUint16(numBytes, bet.number)
	msg = append(msg, numBytes...)

	return msg, nil
}

// TODO: Actualizar para que la agencia solo se mande al principio
func SendBet(conn net.Conn, bet *Bet) error {
	msg, err := encodeBet(bet)
	if err != nil {
		return err
	}
	err = send_all(conn, msg)
	if err != nil {
		return err
	} else {
		return nil
	}
}

// Send a batch of bets to the server
// The first two bytes are the number of bets
// Then it sends the bets
func SendBets(conn net.Conn, bets []*Bet, agency string) error {
	var batches [][]byte
	var currentBatch []byte

	currentBatch = make([]byte, BATCH_SIZE) // Initialize with 2 bytes
	binary.BigEndian.PutUint16(currentBatch, uint16(len(bets)))

	agency_int, _ := strconv.Atoi(agency)
	currentBatch = append(currentBatch, uint8(agency_int))

	for _, bet := range bets {
		betMsg, err := encodeBet(bet)
		if err != nil {
			return err
		}

		// If adding betMsg exceeds the max batch size, flush the current batch
		if len(currentBatch)+len(betMsg) > BATCHMAX {
			batches = append(batches, currentBatch)
			currentBatch = make([]byte, 0)
		}

		currentBatch = append(currentBatch, betMsg...)
	}

	if len(currentBatch) > 0 {
		batches = append(batches, currentBatch)
	}

	for _, batch := range batches {
		if err := send_all(conn, batch); err != nil {
			return err
		}
	}

	return nil
}

func RecvAnswer(conn net.Conn) (int, error) {
	answer, err := RecvAll(conn, ANSWER_SIZE)
	if err != nil {
		return -1, err
	}
	answer_v := int(answer[0])
	return answer_v, nil
}
