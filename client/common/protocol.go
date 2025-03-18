package common

import (
	"encoding/binary"
	"errors"
	"net"
)

const AGENCY_SIZE = 1
const STR_SIZE = 1
const MAX_STR_SIZE = 255
const DOCUMENT_SIZE = 4
const BIRTHDATE_SIZE = 10
const NUMBER_SIZE = 2
const ANSWER_SIZE = 1

// result constants
const SUCESS = 0
const FAIL = 1

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

func SendBet(conn net.Conn, bet *Bet) error {
	// agency
	msg := make([]byte, 0)
	msg = append(msg, bet.agency)

	// firstName
	msg = serializeUnknownString(bet.firstName, msg)
	if msg == nil {
		return errors.New("error serializing first name")
	}

	// lastName
	msg = serializeUnknownString(bet.lastName, msg)
	if msg == nil {
		return errors.New("error serializing last name")
	}

	// document
	docBytes := make([]byte, DOCUMENT_SIZE)
	binary.BigEndian.PutUint32(docBytes, bet.document)
	msg = append(msg, docBytes...)

	// birthDate
	msg = append(msg, []byte(bet.birthDate)...) 

	// number
	numBytes := make([]byte, NUMBER_SIZE)
	binary.BigEndian.PutUint16(numBytes, bet.number)
	msg = append(msg, numBytes...)

	err := send_all(conn, msg)
	if err != nil {
		return err
	} else {
		return nil
	}
}

func RecvAnswer(conn net.Conn) (int, error) {
	answer, err := RecvAll(conn, ANSWER_SIZE)
	if err != nil {
		return -1, err
	}
	answer_v := int(answer[0])
	return answer_v, nil
}
