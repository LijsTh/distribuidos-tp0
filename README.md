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

### EJ6

#### Protocolo de Batch

El protocolo para mandar por batch es el siguiente:

```
| NUMBER_OF_BETS [2] | AGENCY[1] | BET1 | BET2 | BET3 | ...
```

```
BET:
| NAME_N [1]     | NAME   [N]   | SURNAME_N [1] | SURNAME[N] |
| DOCUMENT [4]  | BIRTHDATE [10] | NUMBER [2]   |
```

Primero se manda la cantidad de apuestas que se van a mandar y la agencia. Luego de mandar estos dos datos se empiezan a mandar las encuestas encodificadas igual manera que el ej5 pero sin la agencia para no repetir.

En el caso que el paquete se pase del limite de 8kb se siguen mandando paquetes que no superen el limite con las apuestas restantes haciendo que el mensaje que le llege al servidor y que este mandando el cliente respeta la variable `MaxAmount`. Notar la diferencia entre paquete y mensaje.

También en el struct de apuesta presentado en el ejercicio anterior ya no esta la agencia como atributo.

## EJ7

### Protocolo sorteo

Cuando el cliente termina de mandar las apuestas, envia lo siguiente:

```
| 0 [2] | AGENCY[1] |
```

Manda un cero (2 bytes para q sea compatible con N_BETS del batch) indicando que no hay mas bets para mandar para luego mandar su agency.

Luego el servidor se guarda los clientes que terminaron para realizar el sorteo. Al finalizar el mismo le manda a cada una de las agencias/clientes los ganadores de sus correspondientes agencias con el siguiente formato:

```
N_GANADORES[1] | DOCUMENT_1[4] | DOCUMENT_2[4] | ... | DOCUMENT_N |
```

Finalmente cada cliente le manda un 1 para indicar que recibio los resultados para que servidor termine.

## Ej8

Para este ejercicio se utilizo una pool de procesos en el servidor en la que se matiene un pool del tamaño de las agencias. Cada proceso es encargado de manejar la conexión con el cliente asociado.

### Sincronización

Primero, para la escritura de archivos se utilizo un `Lock` en el que solo un proceso podia estar utilizando el archivo. Este lock es manejado internamente por el `manager` de `multiprocessing`.

Luego para la coordinación del sorteo se utilizo una `List` también manejada por le manager para garantizar acceso seguro, en donde cada proceso se agregaba a la lista una vez terminando. Luego de agregarse a la lista se queda a la espera de un mensaje en una blocking `queue` (también manejada por el `manager`) para saber si es momento del sorteo.

A su vez, cuando un proceso terminaba se fijaba si ya todos habian terminado (el era el último) leyendo el tamaño de la lista luego de agregarse. En el caso de que si sea el último, agrega N cantidad de mensajes como N procesos haya a la `queue` mencionada anteriormente.

De esta manera, cada proceso consume de la `queue` y procede a mandar los resultados de la encuesta sabiendo que al recibir el mensaje de la queue ya estan todos los procesos en el estado de mandar los resultados.
