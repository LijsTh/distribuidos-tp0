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

### EJ4

Para el ejercicio cuatro se utilizo un contexto de go en el cliente para el manejo de señales y un handler de `signal` en python para lo mismo.

De esta manera en el cliente puedo saber cuando el contexto esta terminado (`Done`). Por otro lado en el servidor se maneja con un handler para cerrar los sockets una vez llegada la señal.

### EJ5

#### Protocolo de apuestas

Las apuestas de parte del cliente se modelan con el siguiente struct:

```go
type Bet struct {
	agency uint8
	firstName string
	lastName string
	document uint32
	birthDate string
	number uint16
}
```

El cliente primero le envia un paquete con la apuesta en cuestion serializada de la siguiente manera:

```
| AGENCY   [1]  | NAME_N [1]     | NAME   [N]   | SURNAME_N [1] | SURNAME[N] |
| DOCUMENT [4]  | BIRTHDATE [10] | NUMBER [2]   |
```

Primero se manda el numero de agencia, luego se manda el largo del nombre junto con el mismo. Se repite para el appelido para finalmente enviar el documento, la fecha de nacimiento y el numero del sorteo.

Entre corchetes se encuentra el tamaño en bytes de cada uno de los campos: ex. agencia ocupa un byte. Notar que el valor maximo posible de NAME y SURNAME es 255.

Luego en respuesta el servidor le envia un 0 representando que se guardo la apuesta correctamente o un 1 en caso de que hubo un error.

Finalmente el cliente lee la respuesta del servidor y procede a seguir mandando apuestas.
