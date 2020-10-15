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
	"container/list"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"time"

	pb "helloworld"

	"google.golang.org/grpc"
	//pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address     = "dist29:50051"
	defaultName = "Bro"
	clientName  = "CLIENTES"
)

//Items contiene info acerca de un producto
type Items struct {
	id        string
	name      string
	prioridad string
	data      []string
}

//lee pymes.csv y entrega un hashmap con los productos de pymes con prioridad 0 o 1
func pymeOrders() *list.List {
	// path to csv
	fp, err := os.Open("csv_Files/pymes.csv")
	if err != nil {
		log.Fatalln("Can't open file: ", err)
	}

	r := csv.NewReader(fp)

	itemsPyme := list.New()

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("error reading file: ", err)
		}

		producto := Items{id: record[0], name: record[1], prioridad: record[5], data: []string{record[2], record[3], record[4]}}

		itemsPyme.PushBack(producto)

	}
	return itemsPyme
}

//lee pymes.csv y entrega un hashmap con los productos de pymes con prioridad 0 o 1
func retailOrders() *list.List {
	// path to csv
	fp, err := os.Open("csv_Files/retail.csv")
	if err != nil {
		log.Fatalln("Can't open file: ", err)
	}

	r := csv.NewReader(fp)

	itemsRetail := list.New()

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("error reading file: ", err)
		}

		producto := Items{id: record[0], name: record[1], prioridad: "2", data: []string{record[2], record[3], record[4]}}

		itemsRetail.PushBack(producto)

	}
	return itemsRetail
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

/* hacerOrden - hace pedidos automaticamente parra un cliente
*  p es la lista de la cual hacer el pedido
*  c es la conexion
 */
func hacerOrden(ctx context.Context, p *list.List, c pb.GreeterClient) {

	if p != nil {

		front := p.Front()
		itemI := Items(front.Value.(Items))

		tipo := "none"

		if itemI.prioridad == "0" {
			tipo = "Normal"
		} else if itemI.prioridad == "1" {
			tipo = "prioritario"
		} else {
			tipo = "retail"
		}

		// generacion de orden
		orden := &pb.OrderRequest{
			OrderID:      itemI.id,
			ProductName:  itemI.name,
			ProductValue: itemI.data[0],
			Src:          itemI.data[1],
			Dest:         itemI.data[2],
			Priority:     itemI.prioridad,
			ProductType:  tipo,
		}

		// Hacer una consulta
		r, err := c.MakeOrder(ctx, orden)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}

		// mostrar codigo seguimiento por pantalla
		fmt.Print("\n<--------------- INFORMATION --------------->\n")
		log.Printf("Order Tracking Code: %s\n", r.GetMessage())
		fmt.Println("\n<--------------- INFORMATION --------------->")

		p.Remove(front)
	} else {
		fmt.Println("No hay mas ordenes que enviar.")
	}

}

func main() {
	//--------------------------------------read csv files and store data
	pymes := pymeOrders()
	retails := retailOrders()

	//---------------------------------------

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Contact the server and print out its response.

	//waiting Time
	fmt.Println("\nIngrese Tiempo de Espera en Segundos")
	waitingTime, _ := strconv.Atoi(getInput(2))
	fmt.Printf("\nTiempo: %d\n", waitingTime)

	//Tipo de Cliente
	fmt.Println("\nSeleccione tipo de Cliente\n\n\t[1] Pyme\n\t[2] Retail\n\n\tPara Salir: EXIT")
	opcion := getInput(2)
	if opcion == "EXIT" {
		fmt.Println("Saliendo del programa.")
		os.Exit(3)
	}

	//Thread hacer ordenes
	//soy Pyme
	if opcion == "1" {
		go hacerOrden(ctx, pymes, c)
	} else { //Soy Retail
		go hacerOrden(ctx, retails, c)
	}

	// Codigo para realizer seguimiento

	time.Sleep((5 * time.Second))

	for {
		//Ingresar Codigo de Seguimiento
		fmt.Println("\nIngrese Codigo de Seguimiento\nPara Salir: EXIT")
		code := getInput(3)
		if code == "EXIT" {
			break
		}

		if len(os.Args) > 1 {
			code = os.Args[1]
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		// make tracking request to logistics
		request := &pb.TrackingRequest{
			TrackingCode: code,
		}
		r, err := c.TrackingOrder(ctx, request)
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}

		// show user query results
		fmt.Print("\n<--------------- Status --------------->\n")
		log.Printf("Order Status: %s\n", r.GetMessage())
		fmt.Println("\n<--------------- Status --------------->")
	}
}
