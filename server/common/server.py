import socket
import logging
import signal

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.client = None
        self.running = True

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        signal.signal(signal.SIGINT, self.__shutdown)
        signal.signal(signal.SIGTERM, self.__shutdown)

        while self.running:
            try:
                client_sock = self.__accept_new_connection()
                if client_sock:
                    self.client = client_sock
                    self.__handle_client_connection()

            except OSError as e:
                if self.running:
                    logging.error(f"action: accept_connection | result: fail | error: {e}")
            
            logging.info("action: server_shutdown | result: success")

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            # TODO: Modify the receive to avoid short-reads
            msg = self.client.recv(1024).rstrip().decode('utf-8')
            addr = self.client.getpeername()
            logging.info(f'action: receive_message | result: success | ip: {addr[0]} | msg: {msg}')
            # TODO: Modify the send to avoid short-writes
            self.client.send("{}\n".format(msg).encode('utf-8'))
        except OSError as e:
            if self.running:
                logging.infof("action: closing_client_connection | result: success")
            else:
                logging.error("action: receive_message | result: fail | error: {e}")
        finally:
            if self.client:
                self.client.close()
                self.client = None

    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        c, addr = self._server_socket.accept()
        logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
        return c
    
    def __kill_connections(self):
        if self.client:
            self.client.close()
            logging.info('action: client_connection_shutdown | result: success ')
    
    def __shutdown(self, sig, _):
        logging.info(f"action: {signal.Signals(sig).name} | result: success")
        self.running = False
        self._server_socket.close()
        logging.info(f"action: server_skt_shutdown | result: success")
        self.__kill_connections()

