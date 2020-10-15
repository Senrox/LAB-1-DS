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
	"context"
	"fmt"
	"log"
	"net"
	"strconv"
	"time"

	pb "helloworld"

	"google.golang.org/grpc"
	//pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
	port = ":50051"
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

	trackingCode += 1

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

	//TODO generar codigo para enviar a las colas

	//Codigo de seguimiento para el cliente
	return &pb.OrderConfirmation{Message: strTrackingCode}, nil
}

//consulta de seguimiento a camiones
/*
func (s *server) sendInformation(ctx context.Context, in *pb.Information) (*pb.HelloReply, error) {
	fmt.Println("\n<--------------- INFORMATION STATUS --------------->")
	fmt.Println()

	data := &pb.Information{
		OrderID:      itemI.id,
		ProductType:  itemI.name,
		ProductValue: itemI.data[0],
		Src:          itemI.data[1],
		Dest:         itemI.data[2],
		Attemps:      itemI.prioridad,
		Date:         tipo,
	}

	ProductDatabaseByTracking[in.GetTrackingCode()].status = "New Status" // get status

	store(estado)

	return &pb.HelloReply{Message: "Status" + in.GetName()}, nil
}
*/

// respuesta a consulta de seguimiento
func (s *server) TrackingOrder(ctx context.Context, in *pb.TrackingRequest) (*pb.Status, error) {

	fmt.Println("\n<--------------- STATUS REQUEST --------------->")
	fmt.Println()
	log.Printf("STATUS INFORMATON: %s\n", in.GetTrackingCode())

	//Asignamos estado TEMPORAL
	ProductDatabaseByTracking[in.GetTrackingCode()].status = "New Status" // get status

	//Codigo de respuesta
	return &pb.Status{Message: ProductDatabaseByTracking[in.GetTrackingCode()].status}, nil

}

// ProductDatabaseByTracking - base de datos de tracking
var ProductDatabaseByTracking map[string]*Items

// ProductStatus status de los productos
var ProductStatus map[string]*ItemStatus

//funcion que almacena datos en un hashmap
func store(item Items) {
	ProductDatabaseByTracking[item.tracking] = &item
}

func main() {

	ProductDatabaseByTracking = make(map[string]*Items)

	//-----------------------------------------------------> Server1
	fmt.Printf("Waitin for my bro, date is: %s\n", getTime())
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen1: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve s1: %v", err)
	}

}
