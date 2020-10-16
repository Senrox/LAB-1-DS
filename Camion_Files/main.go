/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
	"context"
	"fmt"
	pb "helloworld"
	"log"
	"math/rand"
	"strconv"
	"time"

	"google.golang.org/grpc"
)

const (
	//address     = ":50051"
	address     = "dist29:50052"
	defaultName = "Bro"
	clientName  = "CAMIONES"
)

//items contiene info acerca de un producto
type Items struct {
	id    string
	tipo  string
	valor string
	src   string
	dest  string
	reply string
	date  string
}

// Envio retorna si el envio se hace o no se hace
func Envio() bool {
	in := []int{0, 1, 1, 1, 1}
	randomIndex := rand.Intn(len(in))
	pick := in[randomIndex]

	if pick == 1 {
		return true
	}
	return false
}

func realizarEnvio(c pb.GreeterClient, tipo string, intentoTime int) {

	// esto dentro del codigo de camiones

	// PEDIR Y RECIBIR UN PAQUETE
	dat := pb.DeliveryRequest{
		R: tipo,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	received, err := c.SendInformation(ctx, &dat)
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}

	var try bool

	var intento int
	var Nintentos int

	var IntentoFinal string
	var enviado bool = false
	var newEstado string = "En Bodega"

	var precio string = received.GetProductValue()
	value, _ := strconv.Atoi(precio)

	if tipo == "retail" {
		fmt.Print("\nRealizo pedido de retail")

		for intento = 0; intento < 3; intento++ {
			//hace cosas

			try = Envio()
			newEstado = "En Camino"
			fmt.Print("\nNuevo estado: En camino")

			if try {
				IntentoFinal = "intento"
				newEstado = "Recibido"
				fmt.Print("\nNuevo estado: Recibido")
				enviado = true
				break
			}
			// ACTUALIZAR ESTADO PAQUETE
			data := &pb.StatusResponse{
				TrackingCode: newEstado,
			}
			received, err := c.TrackingStatus(ctx, data)
			if err != nil {
				log.Fatalf("could not greet: %v%s", err, received)
			}

			// tiempo de espera despues de un envio
			time.Sleep(time.Duration(intentoTime) * time.Second)
		}
		if !try && enviado == false {
			IntentoFinal = "3"
			newEstado = "No Recibido"
			fmt.Print("\nNuevo estado: No Recibido")
		}
	} else { //pyme
		fmt.Print("\nRealizo pedido de pyme")
		if value < 10 {
			Nintentos = 1 // 1 base + 0 extra
		} else if value < 20 {
			Nintentos = 2 // 1 base + 1 extra
		} else { // mayor a 20
			Nintentos = 3 // 1 base + 2 extra
		}

		for intento = 0; intento < Nintentos; intento++ {
			//hace cosas

			try = Envio()
			fmt.Print("\nNuevo estado: En camino")
			newEstado = "En Camino"

			if try {
				IntentoFinal = "intento"
				newEstado = "Recibido"
				fmt.Print("\nNuevo estado: Recibido")
				enviado = true
				break
			}
			// ACTUALIZAR ESTADO PAQUETE
			data := &pb.StatusResponse{
				TrackingCode: newEstado,
			}
			received, err := c.TrackingStatus(ctx, data)
			if err != nil {
				log.Fatalf("could not greet: %v%s", err, received)
			}

			// tiempo de espera despues de un envio
			time.Sleep(time.Duration(intentoTime) * time.Second)
		}
		if !try && enviado == false {
			IntentoFinal = strconv.Itoa(Nintentos)
			newEstado = "No Recibido"
			fmt.Print("\nNuevo estado: No Recibido")
		}

	}

	received.Attempts = IntentoFinal
	fmt.Print("\nNumero de intentos: %d", IntentoFinal)

	orderUpdate := &pb.StatusResponse{
		TrackingCode: received.OrderID,
		Status:       newEstado,
	}
	m, err := c.TrackingStatus(ctx, orderUpdate)
	if err != nil {
		log.Fatalf("could not greet: %v%s", err, m)
	}

}

func camion(c pb.GreeterClient, tipo string, intentoTime int, pedidoTime int) {
	realizarEnvio(c, tipo, intentoTime)
	// tiempo de espera despues de un envio
	time.Sleep(time.Duration(pedidoTime) * time.Second)
	realizarEnvio(c, tipo, intentoTime)
}

//gets input from user
func getInput(x int) string {
	if x == 1 {
		fmt.Print("\nIngrese ID producto a ordenar: ")
	} else if x == 3 {
		fmt.Print("\nIngrese ID producto a Realizar Seguimiento: ")
	} else {
		fmt.Print("\nSeleccion: ")
	}
	var input string
	fmt.Scanln(&input)
	return input
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	//waiting Time 1
	fmt.Println("\nIngrese tiempo de envio entre cada paquete")
	intentoTime, _ := strconv.Atoi(getInput(2))
	fmt.Printf("\nTiempo: %d\n", intentoTime)

	//waiting Time 2
	fmt.Println("\nIngrese tiempo de envio entre cada pedido")
	pedidoTime, _ := strconv.Atoi(getInput(2))
	fmt.Printf("\nTiempo: %d\n", pedidoTime)

	// Contact the server and print out its response.

	for {
		go camion(c, "retail", intentoTime, pedidoTime)
		time.Sleep(3 * time.Second)

		go camion(c, "retail", intentoTime, pedidoTime)
		time.Sleep(3 * time.Second)

		camion(c, "pyme", intentoTime, pedidoTime)
		time.Sleep(3 * time.Second)
	}
}
