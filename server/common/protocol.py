import socket 
from common.utils import Bet


AGENCY_SIZE = 1
STR_SIZE = 1
MAX_STR_SIZE = 255
DOCUMENT_SIZE = 4
BIRTHDATE_SIZE = 10
NUMBER_SIZE = 2
ANSWER_SIZE = 1
BATCH_SIZE = 2

# result constants
SUCCESS = 0
FAIL = 1

# batch max
BATCHMAX = 8000 #8kb


"""
reads the entire message from the socket to avoid short-reads
"""
def recv_all(skt: socket.socket, size: int) -> bytes:
    data = b'' 
    while len(data) < size:
        pkt = skt.recv(size - len(data))
        if len(pkt) == 0:
            raise OSError("Connection closed")
        data += pkt

    return data

"""
sends the entire message to the socket to avoid short-writes
"""
def send_all(skt: socket.socket, data: bytes) -> None: 
    while len(data) > 0:
        sent = skt.send(data)
        if sent == 0:
            raise OSError("Connection closed")
        data = data[sent:]


"""
reads a string from the socket with an unknown size, the size is read first
"""
def recv_str_uknown_size(skt: socket.socket) -> str:
    data = recv_all(skt, STR_SIZE)
    size = int.from_bytes(data, byteorder='big')
    if size > MAX_STR_SIZE:
        raise ValueError("String size exceeded") 
    data = recv_all(skt, size)
    return data.decode('utf-8')


"""
Reads a bet from the socket and returns it.
The bet is expected to be in the following format:
    - 1 byte for the first name size
    - (0-255) bytes for the first name
    - 1 byte for the last name size
    - (0-255) bytes for the last name
    - 4 bytes for the document
    - 10 bytes for the birthdate
    - 2 bytes for the number
"""

def recv_bet(skt: socket.socket, agency: str) -> Bet:
    data = recv_str_uknown_size(skt)
    first_name = data
    data = recv_str_uknown_size(skt)
    last_name = data
    data = recv_all(skt, DOCUMENT_SIZE)
    document = int.from_bytes(data, byteorder='big')
    data = recv_all(skt, BIRTHDATE_SIZE) 
    birthdate = data.decode('utf-8')

    data = recv_all(skt, NUMBER_SIZE)
    number = int.from_bytes(data, byteorder='big')

    return Bet(agency, first_name, last_name, document, birthdate, number)


def recv_batch(skt: socket.socket) -> list[Bet]:
    data = recv_all(skt, BATCH_SIZE)
    size = int.from_bytes(data, byteorder='big')
    agency = recv_all(skt, AGENCY_SIZE)
    agency = int.from_bytes(agency, byteorder='big')
    bets = []
    for _ in range(size):
        bets.append(recv_bet(skt, agency))
    return bets


"""
Send the answer SUCCESS OR FAIL to the client. 
"""
def send_answer(skt: socket.socket, answer: int) -> None:
    if answer not in [SUCCESS, FAIL]:
        raise ValueError("Invalid answer")
    data = answer.to_bytes(ANSWER_SIZE, byteorder='big')
    send_all(skt, data)
