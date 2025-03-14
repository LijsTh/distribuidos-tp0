# TP0: Docker + Comunicaciones + Concurrencia

En este repositorio se encuentra la solución al tp 0 de la materia sistema distribuidos realizada por theo lijs.

## Ejecución

La ejecucción estandard se realiza con

```shell
make docker-compose-up
```

Para finalizar los containers se hace con el comando:

```shell
make docker-compose-down
```

## Resolución

### EJ1

Para generar un yaml con mas clientes se utiliza el script `generar-compose.sh` con los siguientes argumentos:

```shell
./generar-compose.sh <nombre-archivo.yaml> <cantidad clientes>
```
### EJ2

Se utilizo dos volumenes, uno en el servidor y otro en el cliente para poder cumplir con el requisito de no tener que reconstruir las imagenes para que se apliquen los cambios en las configuraciones.

```yaml
server:
        ...

    volumes:
      - ./server/config.ini:/config.ini

client1:
        ...

    volumes:
      - ./client/config.yaml:/config.yaml

```

> Con esta notación los volumenes son de tipo bind.

### EJ3

Se realizo el script `validar-echo-server.sh` que utiliza la network declarada en `docker-compose-dev.yaml` para que un contenedor con la imagen de alpine se pueda comunicar utilizando netcat y de esta manera comprobar que es un echo server. Se decidió utilizar alpine ya que es una imagen ligera con las herramientas necesarias (`nc`)

se utilizan los siguientes comandos para guardarse la respuesta de netcat: 

```shell
    docker run --rm --network "$NETWORK" alpine sh -c " echo $MSG | nc $SERVER $PORT" 
```

> $NETWORK : tp0_testing_net , $SERVER: server , $PORT: 12345. Los valores de las constantes son los defaults en el config.ini del servidor. 




