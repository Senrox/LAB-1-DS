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

// constantes de puertos y nombres de instancias
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

/*
	pymeOrders()
	lee pymes.csv y entrega un hashmap con los productos de pymes con prioridad 0 o 1
	Input: nada
	returns: *list.List itemsPyme, lista de los items de una pyme
*/
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

		if record[0] != "id" {

			producto := Items{id: record[0], name: record[1], prioridad: record[5], data: []string{record[2], record[3], record[4]}}

			itemsPyme.PushBack(producto)
		}

	}
	defer fp.Close()
	return itemsPyme
}

/*
	retailOrders()
	lee pymes.csv y entrega un hashmap con los productos de pymes con prioridad 0 o 1
	Input: nada
	returns: *list.List itemsRetail, lista de los items de un retail
*/
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

		if record[0] != "id" {

			producto := Items{id: record[0], name: record[1], prioridad: "2", data: []string{record[2], record[3], record[4]}}

			itemsRetail.PushBack(producto)
		}

	}
	defer fp.Close()
	return itemsRetail
}

//gets input from user
/*
	getInput()
	obtiene input del usuario y printea texto dependiendo de la opcion escogida
	Input: int x, opcion el usuario
	returns: string input, input que puso el usuario
*/
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

/*
	hacerOrden()
	hace pedidos automaticamente para un cliente desde una lista esperando un intervalo de tiempo
	Input: lista p, conexion c, tiempo de espera waitingTime
	returns: nada
*/
func hacerOrden(p *list.List, c pb.GreeterClient, waitingTime int) {

	for {
		if p.Front() == nil {
			fmt.Println("No hay mas ordenes que enviar.")
			break

		} else {
			// waiting time
			time.Sleep(time.Duration(waitingTime) * time.Second)

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

			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
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

		}
	}
}

func main() {
	//--------------------------------------read csv files and store data
	pymes := pymeOrders()
	retails := retailOrders()

	//---------------------------------------

	// Set up a connection to the server.
	// Contact the server and print out its response.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

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
	if opcion == "1" {
		go hacerOrden(pymes, c, waitingTime)
	} else { //Soy Retail
		go hacerOrden(retails, c, waitingTime)
	}

	// Codigo para realizer seguimiento

	time.Sleep(time.Duration(waitingTime) * time.Second)

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
