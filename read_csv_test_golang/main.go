package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
)

type Items struct {
	id        string
	name      string
	prioridad string
	data      []string
}

var ItemsByID map[string]*Items

func store(item Items) {
	ItemsByID[item.id] = &item
}

// pymer: 	id,producto,valor,tienda,destino,prioritario
// retails: id,producto,valor,tienda,destino
func main() {

	fp, err := os.Open("C:\\Users\\marth\\OneDrive\\Desktop\\2020-2\\Distribuidos\\Lab 1\\Lab1_arch_ej\\pymes.csv")
	if err != nil {
		log.Fatalln("Can't open file: ", err)
	}

	r := csv.NewReader(fp)

	ItemsByID = make(map[string]*Items)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalln("error reading file: ", err)
		}

		producto := Items{id: record[0], name: record[1], prioridad: record[5], data: []string{record[2], record[3], record[4]}}

		store(producto)

	}
	fmt.Println(ItemsByID["GG301"])
	fmt.Println(ItemsByID["CC121"])

}
