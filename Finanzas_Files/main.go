// reciver / consummer
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/streadway/amqp"
)

type Items2 struct {
	Id          string `json:"id"`
	Order_type  string `json:"order_type"`
	Order_value string `json:"order_value"`
	Tracking    string `json:"tracking"`
	Status      string `json:"status"`
	Atts        string `json:"atts"`
}

type Balance struct {
	Id       string
	Tracking string
	Atts     string
	ganancia string
	perdida  string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func guardar() {
	fmt.Print("xd")
}

func main() {
	// Inicaimos conexion
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
		false,         // delete when usused
		false,         // exclusive
		false,         // no-wait
		nil,           // arguments
	)
	failOnError(err, "Failed to declare a queue")

	//se reciven msg del la cola
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	//bloquea la ejecucion del main.go hasta que recibe un valor
	forever := make(chan bool)

	/*
		type Items2 struct {
		Id          string `json:"id"`
		Order_type  string `json:"order_type"`
		Order_value string `json:"order_value"`
		Tracking    string `json:"tracking"`
		Status      string `json:"status"`
		Atts        string `json:"atts"`
		}
	*/
	var gananciasTotal int = 0
	var perdidasTotal int = 0
	var enviosCompletados int = 0
	var enviosFallidos int = 0

	go func() {
		for d := range msgs {
			//log.Printf("Received a message: %s", d.Body)

			var reading Items2

			err = json.Unmarshal([]byte(d), &reading)
			if err != nil {
				log.Fatalf("oh shoiit: %v", err)
			}

			var balance Balance

			balance.Id = reading.Id
			balance.Tracking = reading.Tracking
			balance.Atts = reading.Atts

			intentos, err := strconv.Atoi(reading.Atts)
			tempGanancias, err := strconv.Atoi(reading.Order_value)

			var tempPerdidas int = 10 * intentos
			var ingreso int

			if reading.Status == "Recibido" {
				enviosCompletados++
				ingreso = tempGanancias - tempPerdidas
				if ingreso > 0 {
					gananciasTotal = gananciasTotal + ingreso
					balance.ganancia = strconv.Itoa(ingreso)
					balance.perdida = "0"
				} else {
					perdidasTotal = perdidasTotal + ingreso
					balance.ganancia = "0"
					balance.perdida = strconv.Itoa(ingreso)
				}

			} else { //No Recibido
				if reading.Order_type == "0"
			}

			//fmt.Printf("%s", reading.Id)
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}
