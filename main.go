package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type KeyValue struct {
	key   string
	value string
}
type HashTable struct {
	table [512]*KeyValue
}

func (ht *HashTable) hash(key string) int {
	hash := 0
	for i := 0; i < len(key); i++ {
		hash += int(key[i])
	}
	return hash % 512
}

func (ht *HashTable) hset(key string, value string) string {
	newKeyValue := &KeyValue{key, value}
	index := ht.hash(key)
	if ht.table[index] == nil {
		ht.table[index] = newKeyValue
		return key
	} else {
		if ht.table[index].key == key {
			return key
		} else {
			for i := index; i < 512; i++ {
				if ht.table[i] == nil {
					ht.table[i] = newKeyValue
					return key
				}
			}
		}
	}
	return "Неудолось добавить элемент"
}
func (ht *HashTable) hdel(key string) string {
	index := ht.hash(key)
	if ht.table[index] == nil {
		return "Элемент не найден"
	} else if ht.table[index].key == key {
		ht.table[index] = nil
		return ""
	} else {
		for i := index; i < 512; i++ {
			if ht.table[i].key == key {
				ht.table[i] = nil
				return ""
			}
		}
	}
	return "Неудалось удалить элемент"
}
func (ht *HashTable) hget(key string) string {
	index := ht.hash(key)
	if ht.table[index] == nil {
		return "Элемент не найден"
	} else if ht.table[index].key == key {
		return ht.table[index].value
	} else {
		for i := index; i < 512; i++ {
			if ht.table[i].key == key {
				return ht.table[index].value
			}
		}
	}
	return "Элемент не найден"
}

func (ht *HashTable) readHashFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			file, createErr := os.Create(filename)
			if createErr != nil {
			}
			file.Close()
			return
		}
	}

	lines := strings.Split(string(content), "\n")

	for _, line := range lines {
		parts := strings.Split(line, " ")
		if len(parts) >= 2 {
			key := parts[0]
			value := strings.Join(parts[1:], " ")
			err := ht.hset(key, value)
			if err != "" {
			}
		}
	}
}

func (ht *HashTable) writesHashFile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
	}
	defer file.Close()

	for i := 0; i < 512; i++ {
		if ht.table[i] != nil {
			_, err = file.WriteString(ht.table[i].key + " " + ht.table[i].value + "\n")
			if err != nil {
			}
		}
	}
	return
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		input := scanner.Text()
		args := strings.Fields(input)
		actions := args[0]
		var key string
		var value string
		if len(args) == 2 {
			key = args[1]
			value = ""
		} else if len(args) == 3 {
			key = args[1]
			value = args[2]
		}
		var mut sync.Mutex
		hashTable := &HashTable{}
		hashTable.readHashFile("Url.txt")
		if actions == "HSET" {
			mut.Lock()
			a := hashTable.hset(key, value)
			_, err := conn.Write([]byte(a + "\n"))
			if err != nil {
				fmt.Println("Ошибка при отправке команды на сервер:", err)
				return
			}
		} else if actions == "HGET" {
			mut.Lock()
			a := hashTable.hget(key)
			_, err := conn.Write([]byte(a + "\n"))
			if err != nil {
				fmt.Println("Ошибка при отправке команды на сервер:", err)
				return
			}
		} else if actions == "REPORT" {
			mut.Lock()
			data, err := os.ReadFile("stat/connection.json")
			_, err = conn.Write(data)
			if err != nil {
				continue
			}
		}
		hashTable.writesHashFile("Url.txt")
		mut.Unlock()
	}
}

func main() {
	fmt.Println("Сервер создан")

	ln, err := net.Listen("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Ошибка при запуске сервера:", err)
		return
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Ошибка при принятии соединения:", err)
			continue
		}

		go handleConnection(conn)
	}
}
