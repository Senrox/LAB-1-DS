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
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	pb "helloworld"

	"google.golang.org/grpc"
	//pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	address     = "localhost:50051"
	defaultName = "Bro"
	clientName  = "CLIENTES"
)

//items contiene info acerca de un producto
type Items struct {
	id        string
	name      string
	prioridad string
	data      []string
}

//funcion que almacena datos en un hashmap
func store(dict map[string]*Items, item Items) {
	dict[item.id] = &item
}

//lee pymes.csv y entrega un hashmap con los productos de pymes con prioridad 0 o 1
func priorityOrders() map[string]*Items {
	// path to csv
	fp, err := os.Open("C:\\Users\\marth\\OneDrive\\Desktop\\2020-2\\Distribuidos\\Lab 1\\Lab1_arch_ej\\pymes.csv")
	if err != nil {
		log.Fatalln("Can't open file: ", err)
	}

	r := csv.NewReader(fp)

	ItemsByID := make(map[string]*Items)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("error reading file: ", err)
		}

		producto := Items{id: record[0], name: record[1], prioridad: record[5], data: []string{record[2], record[3], record[4]}}

		store(ItemsByID, producto)

	}
	return ItemsByID
}

//lee retail.csv y entrega un hashmap con los productos de retail 2
func retailOrders(dict map[string]*Items) {
	// path to csv
	fp, err := os.Open("C:\\Users\\marth\\OneDrive\\Desktop\\2020-2\\Distribuidos\\Lab 1\\Lab1_arch_ej\\retail.csv")
	if err != nil {
		log.Fatalln("Can't open file: ", err)
	}

	r := csv.NewReader(fp)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("error reading file: ", err)
		}

		producto := Items{id: record[0], name: record[1], prioridad: "2", data: []string{record[2], record[3], record[4]}}

		store(dict, producto)

	}
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
	//read csv files
	itemsPriority := priorityOrders()
	retailOrders(itemsPriority)

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.

	for {

		fmt.Println("\nSeleccione accion a realizar\n\n\t[1] Hacer Seguimiento\n\t[2] Realizar una Orden\n\n\tPara Salir: EXIT")
		opcion := getInput(2)

		if opcion == "EXIT" {
			break
		}

		//Hacer seguimienteo
		if opcion == "1" {
			// get tracking code from user
			code := getInput(3)
			if !(code != "EXIT") {
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

		} else { //Hacer pedido

			productID := getInput(1)

			if !(productID != "EXIT") {
				break
			}
			if len(os.Args) > 1 {
				productID = os.Args[1]
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			tipo := "none"
			if itemsPriority[productID].prioridad == "0" {
				tipo = "Normal"
			} else if itemsPriority[productID].prioridad == "1" {
				tipo = "prioritario"
			} else {
				tipo = "retail"
			}

			orden := &pb.OrderRequest{
				OrderID:      itemsPriority[productID].id,
				ProductName:  itemsPriority[productID].name,
				ProductValue: itemsPriority[productID].data[0],
				Src:          itemsPriority[productID].data[1],
				Dest:         itemsPriority[productID].data[2],
				Priority:     itemsPriority[productID].prioridad,
				ProductType:  tipo,
			}

			r, err := c.MakeOrder(ctx, orden)
			if err != nil {
				log.Fatalf("could not greet: %v", err)
			}

			// show user query results
			fmt.Print("\n<--------------- INFORMATION --------------->\n")
			log.Printf("Order Tracking Code: %s\n", r.GetMessage())
			fmt.Println("\n<--------------- INFORMATION --------------->")

		}
	}
}
