# Lab-1-Sistemas-Distribuidos
2020-2
______________

## Laboratorio 1
______________


## Integrantes:
    - Campus San Joaquin
    - 201773617-8   Zhuo Chang
    - 201773557-0   Martín Salinas
______________


## Maquinas 
Máquina 1   Logistica --------> Tiene grpc & rabbit
hostname:   dist29
contraseña: BcSz2fUS

Máquina 2   camiones  --------> Tiene grpc
hostname:   dist31
contraseña: jzCsSjfR

Máquina 3   clientes  --------> Tiene grpc
hostname:   dist30
contraseña: CtXTq9qq

Máquina 4   finanzas  --------> Tiene rabbit
hostname:   dist32
contraseña: k5PfFYfP

El usuario de las máquinas es: root
______________


## Instrucciones de uso:

#### Se debe entrar a la carpeta LAB-1-DS
#### Se debe compilar el archivo helloworld.proto dentro de la carpeta A_Dependencies usando el comando:
    protoc -I="." --go_out=$GOROOT/src/helloworld --go-grpc_out=$GOROOT/src/helloworld helloworld.proto

#### Maquina 1:
    - Primero se deben ejecutar los comandos de rabbitmq en la consola: 
            /sbin/service rabbitmq-server start
        para detenerlo, cambiar start por stop
    - luego se debe entrar a la carpeta Logistica_Files
    - Para inicizalizar el servicio de logistica, se debe ejecutar el comando make
    - Para actualizar el repositorio se debe correr el comando make update
    - El comando make clean elimina los archivos csv de esta carpeta

#### Maquina 2:
    - Para inicizalizar el servicio de Camiones, se debe ejecutar el comando make
    - Una vez inicializado el servicio, se deben ingresar los tiempos entre cada intento de envio y cada orden del camion
    - Para actualizar el repositorio se debe correr el comando make update

#### Maquina 3:
    - Para inicizalizar el servicio de Clientes, se debe ejecutar el comando make
    - Una vez inicializado el servicio, se debe ingresar el tiempo entre ordenes
    - Despues de haber ingresado el tiempo entre ordenes, se debe seleccionar el tipo de cliente
    - Para actualizar el repositorio se debe correr el comando make update
    - El comando make clean elimina los archivos csv de esta carpeta

#### Maquina 4:
    - Primero se deben ejecutar los comandos de rabbitmq en la consola: 
            /sbin/service rabbitmq-server start
        para detenerlo, cambiar start por stop
    - Para inicizalizar el servicio de finanzas, se debe ejecutar el comando make
    - Para actualizar el repositorio se debe correr el comando make update
______________


## Consideraciones:
    - Se asume que los archivos csv que contienen las ordenes se llaman pymes.csv y retail.csv
        Ambos estan ubicados dentro de la maquina de clientes en la carpeta LAB-1-DS/Cliente_Files/csv_files
    - Se tomo cada envio del paquete como un intento, por lo tanto el primer intento tiene un costo de 0 y los posteriores tienen un costo de 10.
    - Se asume que el usuario sabe lo que tiene que hacer y no comete errores en sus inputs
    - Se asume que existe una carpeta llamada helloworld dentro de $GOROOT/src
    - Se asume que las variables de entorno estan correctamente actualizadas ($GOROOT, $GOPATH, $GOBIN) en el archivo, .bashrc ubicado en ~/ con lo siguiente
            export GOROOT=/usr/local/go
            export GOPATH=$HOME/go
            export GOBIN=$GOPATH/bin
            export PATH=$PATH:$GOROOT:$GOPATH:$GOBIN
