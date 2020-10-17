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

// Package main implements a server for Greeter service.
package main

import (
	"container/list"
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	pb "helloworld"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	//pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	portClientes = ":50051"
	portCamiones = ":50052"
	portFinazas  = ":50053"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// Contiene info de un producto
type Items struct {
	id          string
	name        string
	order_type  string
	order_dest  string
	order_src   string
	order_value string
	tracking    string
	status      string
	timestamp   string
}

type ItemStatus struct {
	trackingCode string
	status       string
}

// ProductDatabaseByTracking - base de datos de tracking
var ProductDatabaseByTracking map[string]*Items

var ordenesCompletas = list.New()
var colaPrioritario = list.New()
var colaNormal = list.New()
var colaRetail = list.New()

//funcion que retorna el tiempo actual
func getTime() string {
	t := time.Now()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v. From: %v", in.GetName(), in.GetClientName())
	return &pb.HelloReply{Message: "Sup " + in.GetName()}, nil
}

var trackingCode int = 1000

// Server recibe orden y se evnia codigo de seguimiento
func (s *server) MakeOrder(ctx context.Context, in *pb.OrderRequest) (*pb.OrderConfirmation, error) {

	fmt.Println("\n<--------------- NEW ORDER COMES IN!!! --------------->")
	log.Printf("ORDER INFORMATON:")

	trackingCode++

	strTrackingCode := strconv.Itoa(trackingCode)

	// timestamp id-paquete tipo nombre valor origen destino seguimiento
	//fmt.Printf("TIMESTAMP | ORDER ID | TYPE | NAME | VALUE | SOURCE | DEST | TRACKING CODE")
	fmt.Println()
	log.Printf("%s %s %s %s %s %s %d\n",
		in.GetOrderID(), in.GetProductType(), in.GetProductName(),
		in.GetProductValue(), in.GetSrc(), in.GetDest(), trackingCode)

	//store data en el dic del servidor
	orden := Items{
		id:          in.GetOrderID(),
		name:        in.GetProductName(),
		order_type:  in.GetProductType(),
		order_dest:  in.GetDest(),
		order_src:   in.GetSrc(),
		order_value: in.GetProductValue(),
		tracking:    strTrackingCode,
		status:      "En la Cola",
		timestamp:   getTime(),
	}

	store(orden)
	//generar codigo para enviar a las colas

	storeInList(orden)

	//Codigo de seguimiento para el cliente
	return &pb.OrderConfirmation{Message: strTrackingCode}, nil
}

//consulta de seguimiento a camiones

func (s *server) SendInformation(ctx context.Context, in *pb.DeliveryRequest) (*pb.Information, error) {

	fmt.Println("\n<--------------- INFORMATION STATUS --------------->")

	tipoCamion := in.GetR()
	fmt.Printf("Tipo Camion: %s\n", tipoCamion)

	/*
		front := l.Front()
		itemI := Items(front.Value.(Items))

		do stuff

		l.Remove(front)
	*/
	var itemI Items
	var itemII Items
	var flag = true

	if tipoCamion == "retail" {
		if colaRetail.Front() != nil {
			front := colaRetail.Front()
			itemI = Items(front.Value.(Items))
			itemII = itemI
			colaRetail.Remove(front)

		} else if colaPrioritario.Front() != nil {
			front := colaPrioritario.Front()
			itemI = Items(front.Value.(Items))
			itemII = itemI
			colaPrioritario.Remove(front)

		} else {
			fmt.Print("No hay entregas para realizar")
			flag = false
		}
	} else {
		if colaPrioritario.Front() != nil {
			front := colaPrioritario.Front()
			itemI = Items(front.Value.(Items))
			itemII = itemI
			colaPrioritario.Remove(front)

		} else if colaNormal.Front() != nil {
			front := colaNormal.Front()
			itemI = Items(front.Value.(Items))
			itemII = itemI
			colaNormal.Remove(front)

		} else {
			fmt.Print("No hay entregas para realizar")
			flag = false
		}
	}

	/*Items
	type Items struct {
		id          string
		name        string
		order_type  string
		order_dest  string
		order_src   string
		order_value string
		tracking    string
		status      string
		timestamp   string
	}*/
	if flag {
		return &pb.Information{
			OrderID:      itemII.tracking,
			ProductType:  itemII.order_type,
			ProductValue: itemII.order_value,
			Src:          itemII.order_src,
			Dest:         itemII.order_dest,
			Attempts:     "", //Se modifica en la otra func, realizar envio,
			Date:         getTime(),
		}, nil
	}
	var str string
	return &pb.Information{
		OrderID:      str,
		ProductType:  str,
		ProductValue: str,
		Src:          str,
		Dest:         str,
		Attempts:     str,
		Date:         str,
	}, nil
}

// respuesta a consulta de seguimiento
// cliente <-> servidor
func (s *server) TrackingOrder(ctx context.Context, in *pb.TrackingRequest) (*pb.Status, error) {

	fmt.Println("\n<--------------- STATUS REQUEST --------------->")
	fmt.Println()
	log.Printf("STATUS INFORMATON: %s\n", in.GetTrackingCode())

	//Codigo de respuesta
	return &pb.Status{Message: ProductDatabaseByTracking[in.GetTrackingCode()].status}, nil
}

// actualizacion de estados
// camoines <-> servidor
func (s *server) TrackingStatusUpdate(ctx context.Context, in *pb.StatusResponse) (*pb.MsgGenerico, error) {

	ProductDatabaseByTracking[in.GetTrackingCode()].status = in.GetStatus()

	fmt.Println("\n<--------------- STATUS UPDATE --------------->")
	fmt.Println()
	log.Printf("Tracking Code: %s\n", in.GetTrackingCode())
	log.Printf("STATUS INFORMATON: %s\n", in.GetStatus())

	//Codigo de respuesta
	var str string
	return &pb.MsgGenerico{Message: str}, nil
}

func (s *server) TrackingStatusFinal(ctx context.Context, in *pb.StatusResponse) (*pb.HelloReply, error) {

	ProductDatabaseByTracking[in.GetTrackingCode()].status = in.GetStatus()

	fmt.Println("\n<--------------- STATUS UPDATE --------------->")
	fmt.Println()
	log.Printf("Tracking Code: %s\n", in.GetTrackingCode())
	log.Printf("STATUS INFORMATON: %s\n", in.GetStatus())

	ordenesCompletas.PushBack(in.GetTrackingCode())

	//Codigo de respuesta
	var str string
	return &pb.HelloReply{Message: str}, nil
}

// ProductStatus status de los productos
var ProductStatus map[string]*ItemStatus

//funcion que almacena datos en un hashmap
func store(item Items) {
	ProductDatabaseByTracking[item.tracking] = &item
}

//funcion que almacena datos en un hashmap
func storeInList(item Items) {
	if item.order_type == "0" {
		colaNormal.PushBack(item)
	} else if item.order_type == "1" {
		colaPrioritario.PushBack(item)
	} else {
		colaRetail.PushBack(item)
	}
}

func clientes() {
	//--------------------------------------------------------------> Server1
	fmt.Print("Waitin for my CLientes amigos")
	lis, err := net.Listen("tcp", portClientes)
	if err != nil {
		log.Fatalf("failed to listen1: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve s1: %v", err)
	}
}

func camiones() {
	//--------------------------------------------------------------> Server1
	fmt.Print("Waitin for my trucks, I'm the mothafucka T.R.U.C.K.")
	lis, err := net.Listen("tcp", portCamiones)
	if err != nil {
		log.Fatalf("failed to listen2: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve s2: %v", err)
	}
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func enviarAfinanzas() {
	// se crea conecxion
	conn, err := amqp.Dial("amqp://test:test@10.6.40.169:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// se abre el canal
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// creacion de cola
	q, err := ch.QueueDeclare(
		"hello-queue", // name
		false,         // durable
		false,         // delete when unused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")
	var code string
	for {
		if ordenesCompletas.Front() != nil {

			front := ordenesCompletas.Front()
			code = string(front.Value.(string))

			data := ProductDatabaseByTracking[code]
			fmt.Println(data)
			// envia info
			//Creacion de msg a publicar
			body := "{name:arvind, message:hello}"
			err = ch.Publish(
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        []byte(body),
				})
			log.Printf(" [x] Sent %s", body)
			failOnError(err, "Failed to publish a message")

			ordenesCompletas.Remove(front)
		} else {
			fmt.Print("No hay ordenes completadas")
		}
	}
}

func main() {

	ProductDatabaseByTracking = make(map[string]*Items)

	go camiones()

	clientes()

}
