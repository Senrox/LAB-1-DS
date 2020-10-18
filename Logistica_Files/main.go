// Package main implements a server for Greeter service.
package main

import (
	"container/list"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	pb "helloworld"

	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	//pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

// constantes de los puertos
const (
	portClientes = ":50051"
	portCamiones = ":50052"
	portFinazas  = ":50053"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// Items
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
	atts        string
}

// Items2
// Struct para formatear el struct item y enviarlo como json
type Items2 struct {
	Id          string `json:"id"`
	Order_type  string `json:"order_type"`
	Order_value string `json:"order_value"`
	Tracking    string `json:"tracking"`
	Status      string `json:"status"`
	Atts        string `json:"atts"`
}

// ItemStatus
// struct para guardar paquetes segun su codigo de envio
type ItemStatus struct {
	trackingCode string
	status       string
}

// ProductDatabaseByTracking - base de datos de tracking y datos de cada paquete
var ProductDatabaseByTracking map[string]*Items

//Colas a utilizar
var ordenesCompletas = list.New()
var colaPrioritario = list.New()
var colaNormal = list.New()
var colaRetail = list.New()

// variable para generar codigos de seguimiento
var trackingCode int = 1000

// ProductStatus status de los productos
var ProductStatus map[string]*ItemStatus

/*
	getTime()
	Obtiene el tiempo de la maquina
	Input: no tiene
	returns: fecha actual, string
*/
func getTime() string {
	t := time.Now()
	return fmt.Sprintf("%d-%02d-%02d %02d:%02d:%02d",
		t.Year(), t.Month(), t.Day(),
		t.Hour(), t.Minute(), t.Second())
}

/*
	Sayhello()
	recibe un mensaje de saludo y envia una respuesta
	Input: context, mensaje tipo helloRequest
	returns: mensaje tipo helloreply
*/
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v. From: %v", in.GetName(), in.GetClientName())
	return &pb.HelloReply{Message: "Sup " + in.GetName()}, nil
}

/*
	MakeOrder()
	recibe una order y la procesa retornando la orden
	Input: context, mensaje tipo OrderRequest
	returns: mensaje tipo orderConfirmation
*/
func (s *server) MakeOrder(ctx context.Context, in *pb.OrderRequest) (*pb.OrderConfirmation, error) {

	fmt.Println("\n<--------------- NEW ORDER COMES IN!!! --------------->")
	log.Printf("\tORDER INFORMATON:")
	fmt.Println("\n<--------------- INFORMATION STATUS --------------->")

	trackingCode++

	strTrackingCode := strconv.Itoa(trackingCode)

	// timestamp id-paquete tipo nombre valor origen destino seguimiento
	//fmt.Printf("TIMESTAMP | ORDER ID | TYPE | NAME | VALUE | SOURCE | DEST | TRACKING CODE")
	fmt.Println()
	log.Printf("\t%s %s %s %s %s %s %d\n",
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

	if orden.order_type == "Retail" {
		orden.tracking = "0"
	}

	store(orden)
	//generar codigo para enviar a las colas

	storeInList(orden)

	//Codigo de seguimiento para el cliente
	return &pb.OrderConfirmation{Message: strTrackingCode}, nil
}

/*
	SendInformation()
	recibe un delivery request de un camion y retorna la informacion de un envio segun corresponda
	Input: context, mensaje tipo deliveryRequest
	returns: mensaje de tipo Information
*/
func (s *server) SendInformation(ctx context.Context, in *pb.DeliveryRequest) (*pb.Information, error) {

	fmt.Println("\n<--------------- INFORMATION STATUS --------------->")

	tipoCamion := in.GetR()
	fmt.Printf("\n\tTipo Camion: %s\n", tipoCamion)
	fmt.Println("\n<--------------- INFORMATION STATUS --------------->")

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
			fmt.Println("\n<--------------- SYSTEM UPDATE --------------->")
			fmt.Print("\n\tNo hay entregas para realizar")
			fmt.Println("\n<--------------- SYSTEM UPDATE --------------->")

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
			fmt.Println("\n<--------------- SYSTEM UPDATE --------------->")
			fmt.Print("\n\tNo hay entregas para realizar\n")
			fmt.Println("\n<--------------- SYSTEM UPDATE --------------->")
			flag = false
		}
	}

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

/*
	TrackingOrder()
	recibe un tracking request de un cliente y retorna la informacion del paquete segun corresponda
	Input: context, mensaje tipo trackingRequest
	returns: mensaje de tipo status
*/
func (s *server) TrackingOrder(ctx context.Context, in *pb.TrackingRequest) (*pb.Status, error) {

	fmt.Println("\n<--------------- STATUS REQUEST --------------->")
	log.Printf("\n\tSTATUS INFORMATON: %s\n", in.GetTrackingCode())
	fmt.Println("\n<--------------- STATUS UPDATE --------------->")

	//Codigo de respuesta
	return &pb.Status{Message: ProductDatabaseByTracking[in.GetTrackingCode()].status}, nil
}

/*
	TrackingStatusUpdate()
	recibe una actualizacion de un paquete por parte de un camion y hace el update correspondiente
	Input: context, mensaje tipo statusResponse
	returns: mensaje generico sin importancia
*/
func (s *server) TrackingStatusUpdate(ctx context.Context, in *pb.StatusResponse) (*pb.MsgGenerico, error) {

	ProductDatabaseByTracking[in.GetTrackingCode()].status = in.GetStatus()
	ProductDatabaseByTracking[in.GetTrackingCode()].atts = in.GetAttempts()

	fmt.Println("\n<--------------- STATUS UPDATE --------------->")
	log.Printf("\n\tTracking Code: %s\n", in.GetTrackingCode())
	log.Printf("\n\tSTATUS INFORMATON: %s\n", in.GetStatus())
	fmt.Println("\n<--------------- STATUS UPDATE --------------->")

	//Codigo de respuesta
	var str string
	return &pb.MsgGenerico{Message: str}, nil
}

/*
	TrackingStatusFinal()
	recibe una actualizacion de un paquete por parte de un camion y hace el update correspondiente para el estado final del paquete
	Input: context, mensaje tipo statusResponse
	returns: mensaje de tipo helloReply
*/
func (s *server) TrackingStatusFinal(ctx context.Context, in *pb.StatusResponse) (*pb.HelloReply, error) {

	ProductDatabaseByTracking[in.GetTrackingCode()].status = in.GetStatus()
	ProductDatabaseByTracking[in.GetTrackingCode()].atts = in.GetAttempts()

	fmt.Println("\n<--------------- STATUS UPDATE --------------->")
	log.Printf("\n\tTracking Code: %s\n", in.GetTrackingCode())
	log.Printf("\n\tSTATUS INFORMATON: %s\n", in.GetStatus())
	fmt.Println("\n<--------------- STATUS UPDATE --------------->")

	ordenesCompletas.PushBack(in.GetTrackingCode())

	//Codigo de respuesta
	var str string
	return &pb.HelloReply{Message: str}, nil
}

/*
	Store()
	recibe un struct item y lo guarda en el map segun su codigo de seguimiento
	Input: struct de tipo items
	returns: nada
*/
func store(item Items) {
	ProductDatabaseByTracking[item.tracking] = &item
}

/*
	StoreInList()
	recibe un struct item y lo guarda en la cola de camion que le corresponda
	Input: struct de tipo items
	returns: nada
*/
func storeInList(item Items) {
	if item.order_type == "0" {
		colaNormal.PushBack(item)
	} else if item.order_type == "1" {
		colaPrioritario.PushBack(item)
	} else {
		colaRetail.PushBack(item)
	}
}

/*
	Clientes()
	Conexion con clientes
	Input: nada
	returns: nada
*/
func clientes() {
	//--------------------------------------------------------------> Server1
	fmt.Print("Waitin for my Clientes...")
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

/*
	camiones()
	Conexion con Camiones
	Input: nada
	returns: nada
*/
func camiones() {
	//--------------------------------------------------------------> Server1
	fmt.Print("Waitin for Trucks...")
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

/*
	failOnError()
	comprueba un mensaje de error y lo muestra
	Input: error, string
	returns: nada
*/
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

/*
	enviarAfinanzas()
	Se conecta con finanzas y hace el publish de structs items2 con rabbitmq, ademas hace el registro de paquetes
	Input: f *os.File, archivo a actualizar
	returns: nada
*/
func enviarAfinanzas(f *os.File) {
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

			item := ProductDatabaseByTracking[code]

			item2 := Items2{Id: item.id, Order_type: item.order_type, Order_value: item.order_value, Tracking: item.tracking, Status: item.status, Atts: item.atts}

			byteArray, err := json.Marshal(item2)
			if err != nil {
				fmt.Println(err)
			}

			// envia info
			//Creacion de msg a publicar
			err = ch.Publish(
				"",     // exchange
				q.Name, // routing key
				false,  // mandatory
				false,  // immediate
				amqp.Publishing{
					ContentType: "application/json",
					Body:        byteArray,
				})
			log.Printf(" [x] Sent %s", byteArray)
			failOnError(err, "Failed to publish a message")

			toFile := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s\n", getTime(), item.id, item.order_type, item.name,
				item.order_value, item.order_src, item.order_dest, item.tracking)

			_, err = f.WriteString(toFile)
			check(err)

			ordenesCompletas.Remove(front)

		} else {
			time.Sleep(10 * time.Second)
			fmt.Println("\n<--------------- SYSTEM UPDATE --------------->")
			fmt.Print("\n\tNo hay ordenes completadas")
			fmt.Println("\n<--------------- SYSTEM UPDATE --------------->")

		}
	}
}

/*
	check()
	comprueba un error y lo muestra
	Input: error e
	returns: nada
*/
func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {

	f, err := os.Create("registry.csv")
	check(err)

	ProductDatabaseByTracking = make(map[string]*Items)

	go camiones()

	go enviarAfinanzas(f)

	clientes()

	defer f.Close()
}
