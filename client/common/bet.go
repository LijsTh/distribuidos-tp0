package common

import (
	"os"
	"strconv"
)

type Bet struct {
	agency    uint8
	firstName string
	lastName  string
	document  uint32
	birthDate string
	number    uint16
}

// NewBet Creates a new bet
func NewBet(agency string, firstName string, lastName string, document uint32, birthDate string, number uint16) *Bet {
	agency_n, _ := strconv.Atoi(agency)
	bet := new(Bet)
	bet.agency = uint8(agency_n)
	bet.firstName = firstName
	bet.lastName = lastName
	bet.document = document
	bet.birthDate = birthDate
	bet.number = number
	return bet
}

func NewBetFromEnv() (*Bet, error) {
	agency := os.Getenv("AGENCY")
	firstName := os.Getenv("FIRSTNAME")
	lastName := os.Getenv("LASTNAME")
	document, err := strconv.Atoi(os.Getenv("DOCUMENT"))
	if err != nil {
		return nil, err
	}
	birthDate := os.Getenv("BIRTHDATE")
	number, err := strconv.Atoi(os.Getenv("NUMBER"))
	if err != nil {
		return nil, err
	}
	return NewBet(agency, firstName, lastName, uint32(document), birthDate, uint16(number)), nil
}
