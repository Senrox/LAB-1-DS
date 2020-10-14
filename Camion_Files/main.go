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
	"log"
	"os"
	"time"

	pb "helloworld"

	"google.golang.org/grpc"
)

const (
	//address     = ":50051"
	address     = "10.6.40.169:50051"
	defaultName = "Bro"
	clientName  = "CAMIONES"
)

func getInput() string {
	fmt.Println("Inserte nombre: ")
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

	// Contact the server and print out its response.

	for {

		name := getInput()

		if !(name != "EXIT") {
			break
		}

		if len(os.Args) > 1 {
			name = os.Args[1]
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name, ClientName: clientName})
		if err != nil {
			log.Fatalf("could not greet: %v", err)
		}
		log.Printf("Salutation: %s", r.GetMessage())
	}

}
