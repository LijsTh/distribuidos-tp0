import socket
import logging
import signal
import multiprocessing
from common.utils import Bet, store_bets, load_bets, has_won
from common.protocol import recv_batch, send_answer, send_results, SUCCESS, FAIL


class Server:
    def __init__(self, port, listen_backlog, max_clients):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.max_clients = max_clients
        self.running = True
        self.pool = multiprocessing.Pool()
        self.current_clients = 0

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """
        manager = multiprocessing.Manager()
        file_lock = manager.Lock()
        agencies_done_lock = manager.Lock()
        agencies_done = manager.list()
        lottery_queue = manager.Queue()
        running = manager.Value('b', True)

        signal.signal(signal.SIGINT, lambda s,f: self.__shutdown(s,f,lottery_queue, agencies_done, running))
        signal.signal(signal.SIGTERM, lambda s,f: self.__shutdown(s,f,lottery_queue, agencies_done, running))

        with manager:
            while self.current_clients < self.max_clients:
                try:
                    client_sock = self.__accept_new_connection()
                    if client_sock:
                        self.current_clients += 1
                        self.pool.apply_async(handle_client, args=(client_sock, file_lock, agencies_done_lock, agencies_done, self.max_clients, lottery_queue, running))

                except socket.timeout:
                    logging.info("action: server_loop | result: success| listener: timeout")
                    break
    
                except OSError as e:
                    if self.running:
                        logging.error(f"action: server_loop | result: fail | error: {e}")
                        break
                except Exception as e:
                    logging.error(f"action: server_loop | result: fail | error: {e}")
                    break

            self.__exit(lottery_queue, running)
            
        logging.info("action: server_shutdown | result: success")


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
    
    def __kill_waiting_clients(self, queue, agencies_done):
        for agency in agencies_done:
            logging.debug(f"action: shutting_down_client: | agency: {agency}")
            queue.put(False)


    def __shutdown(self, sig, _, queue, agencies_done, running):
        if sig != None:
            logging.info(f"action: {signal.Signals(sig).name} | result: success")
        running.value = False
        self._server_socket.shutdown(socket.SHUT_RDWR)
        self._server_socket.close()
        self.__kill_waiting_clients(queue, agencies_done)
        logging.info(f"action: server_skt_shutdown | result: success")

    def __exit(self, lottery_queue, running):
            self.pool.close()
            self.pool.join()
            if running.value:
                self.__shutdown(None, None, lottery_queue, [], running)
            while not lottery_queue.empty():
                logging.debug("Emptying leftovers")
                lottery_queue.get()


def handle_client (client, file_lock, agency_lock, agencies_done, max_agencies, lottery_queue, running):
    """
    Read message from a specific client socket and closes the socket

    If a problem arises in the communication with the client, the
    client socket will also be closed

    If the client sends a batch of bets, the bets will be stored in a file. 

    Else if the client sends an empty batch, the client will be marked as finished waiting
    for the lottery to start. The lottery will start when all clients are finished and it will be announced with a TRUE message in the lottery_queue. If a signal is received to shutdown the queue will contain a FALSE message for the process to exit.

    """
    try:
        while True:
            bets = []
            agency = None
            bets, agency = recv_batch(client)
            if len(bets) == 0:
                update_finished_clients(max_agencies, agency_lock, agencies_done, agency, lottery_queue)
                if not lottery_queue.get():
                    return
                bets = []
                with file_lock:
                    bets = load_bets()
                winners = [(bet.agency, int(bet.document)) for bet in bets if has_won(bet)]
                send_results({agency: client}, winners)
                break
            else :
                with file_lock:
                    store_bets(bets)
                logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
                send_answer(client, SUCCESS)

    except OSError:
        logging.info("action: client_closed| result: success")

    except Exception: 
        logging.error(f"action: apuesta_recibida| result: fail | cantidad: {len(bets)}")
        send_answer(client, FAIL)

    finally:
        client.close()

def update_finished_clients(max_agencies, agency_lock, agencies_done, agency, lottery_queue):
    with agency_lock: 
        agencies_done.append(agency)
        logging.info(f"action: cliente_finalizado | result: success | totales: {len(agencies_done)}")
        if len(agencies_done) == max_agencies:
            logging.info("action: sorteo | result: success")
            for agency in agencies_done:
                lottery_queue.put(True)
    
