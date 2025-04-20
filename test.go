package main

import (
	"log"
	"sync"
)

type Document struct {
	Fields map[string]string
}

func main() {
	var documentPool = sync.Pool{
		New: func() interface{} {
			return &Document{Fields: make(map[string]string)}
		},
	}

	// Получение объекта из пула
	doc := documentPool.Get().(*Document)
	// Возврат объекта в пул после использования
	documentPool.Put(doc)

	// Вместо var data []string
	data := make([]string, 0, 1000) // Резервирование памяти под 1000 элементов

	log.Println(data)
}
