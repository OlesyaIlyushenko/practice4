package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Подключено к серверу статистики")

	conn, err := net.Dial("tcp", "localhost:1337")
	if err != nil {
		fmt.Println("Ошибка при подключении к серверу:", err)
		os.Exit(1)
	}
	defer conn.Close()

	stdinScanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("Пример детализации: SourceIP TimeInterval URL")
		fmt.Print("Введите иерархию для детализации: ")
		if stdinScanner.Scan() {
			input := stdinScanner.Text()

			_, err := conn.Write([]byte("2 " + input + "\n"))
			if err != nil {
				fmt.Println("Ошибка при отправке команды на сервер:", err)
				return
			}
		}

		if stdinScanner.Err() != nil {
			fmt.Println("Ошибка при чтении команды:", stdinScanner.Err())
			break
		}
	}
}
