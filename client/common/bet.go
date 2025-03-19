package common

import (
	"encoding/csv"
	"os"
	"strconv"
)

const FILEPATH = "/data/agency-"

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

// Santiago Lionel,Lorca,30904465,1999-03-17,2201
const NOMBRE = 0
const APELLIDO = 1
const DNI = 2
const FECHA_NACIMIENTO = 3
const APUESTA = 4

type BetReader struct {
	agency     string
	finished   bool
	reader     *csv.Reader
	file       *os.File
	batch_size int
}

func NewBetReader(batch_size int, agency string) *BetReader {
	/// Crates a new BetReader
	//  Opens the file ready to read
	path := FILEPATH + agency + ".csv"
	file, _ := os.Open(path)

	reader := csv.NewReader(file)
	bet_reader := new(BetReader)
	bet_reader.reader = reader
	bet_reader.finished = false
	bet_reader.batch_size = batch_size
	bet_reader.file = file
	bet_reader.agency = agency

	return bet_reader
}

func (br *BetReader) ReadBet() (*Bet, error) {
	// Reads a bet from the csv file with the following format:
	// firstName, lastName, document, birthDate, number
	if br.finished {
		return nil, nil
	}
	record, err := br.reader.Read()
	if err != nil {
		if err.Error() == "EOF" {
			br.finished = true
			return nil, nil
		} else {
			return nil, err
		}
	}

	firstName := record[0]
	lastName := record[1]
	document, err := strconv.Atoi(record[2])
	if err != nil {
		return nil, err
	}
	birthDate := record[3]
	number, err := strconv.Atoi(record[4])
	if err != nil {
		return nil, err
	}

	return NewBet(br.agency, firstName, lastName, uint32(document), birthDate, uint16(number)), nil
}

func (br *BetReader) Finished() bool {
	return br.finished
}

func (br *BetReader) ReadBets() ([]*Bet, error) {
	// Reads a batch of bets from the csv file
	var bets []*Bet
	for i := 0; i < br.batch_size; i++ {
		bet, err := br.ReadBet()
		if err != nil {
			return nil, err
		}
		if bet == nil {
			break
		}
		bets = append(bets, bet)
	}
	return bets, nil
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
