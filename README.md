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



