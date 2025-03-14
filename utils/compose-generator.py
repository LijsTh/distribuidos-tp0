import sys

INIT_STR = """name: tp0
services:
"""

SERVER_STR = """  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    environment:
      - PYTHONUNBUFFERED=1
      - LOGGING_LEVEL=DEBUG
    networks:
      - testing_net

"""

NETWORK_STR = """networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24
"""

def client_generate(n):
    client_string = ""
    for i in range(1,int(n)+1):
        client_string += f"""  client{i}:
    container_name: client{i}
    image: client:latest
    entrypoint: /client
    environment:
      - CLI_ID={i}
      - CLI_LOG_LEVEL=DEBUG
    networks:
      - testing_net
    depends_on:
      - server

"""
    return client_string
            


def generate_compose(path, clients):
    yaml_content = INIT_STR + SERVER_STR + client_generate(clients) + NETWORK_STR
    with open(path, "w") as yaml_file:
        yaml_file.write(yaml_content)





if __name__ == "__main__":
    if len(sys.argv) != 3:
        print(" run as: ./compose-generator.py <file.yaml> <clients amount>")
    else:
        path = sys.argv[1]
        clients = sys.argv[2]
        generate_compose(path, clients)

