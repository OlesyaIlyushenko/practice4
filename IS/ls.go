package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
)

var mu sync.Mutex

func main() {
	http.HandleFunc("/shortUrl", shortenHandler)
	http.HandleFunc("/", redirectHandler)
	http.ListenAndServe(":4343", nil)
}


func isValidUrl(token string) bool {
	_, err := url.ParseRequestURI(token)
	if err != nil {
		return false
	}
	u, err := url.Parse(token)
	if err != nil || u.Host == "" {
		return false
	}
	return true
}


func shortenHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.ParseForm()
	originalURL := r.Form.Get("url")

	if originalURL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	conn, err := net.Dial("tcp", "localhost:6379")
	if err != nil {
		fmt.Println("Ошибка при подключении к серверу:", err)
		os.Exit(1)
	}
	defer conn.Close()
	mu.Lock()
	defer mu.Unlock()

	shortURL := generateShortURL(originalURL)
	_, err = conn.Write([]byte("HSET " + shortURL + " " + originalURL + "\n"))
	if err != nil {
		fmt.Println("Ошибка при отправке команды на сервер:", err)
		return
	}
	err_conn, er := bufio.NewReader(conn).ReadString('\n')
	if er != nil {
		fmt.Println("Ошибка при чтении ответа от сервера:", er)
		return
	}
	shortURL = err_conn
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintf(w, "Shortened URL: http://localhost:4343/%s", shortURL)
}


func generateShortURL(input string) string {
	hash := sha256.New()
	hash.Write([]byte(input))
	shortLink := hex.EncodeToString(hash.Sum(nil))

	return shortLink[:7]
}


func redirectHandler(w http.ResponseWriter, r *http.Request) {
	var mut sync.Mutex

	conn, err := net.Dial("tcp", ":6379")
	if err != nil {
		fmt.Println("Ошибка при подключении к серверу")
		os.Exit(1)
	}
	defer conn.Close()
	mut.Lock()
	defer mut.Unlock()

	shortUrl := strings.TrimPrefix(r.URL.Path, "/")
	_, err = conn.Write([]byte("HGET " + shortUrl + "\n"))
	if err != nil {
		fmt.Println("Ошибка при отправки на сервер", err)
		os.Exit(1)
	}
	// Читаем ответ от сервера
	original, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Ошибка при чтении с сервера")
		os.Exit(1)
	}

	if original != "Элемент не найден" {
		http.Redirect(w, r, original, http.StatusFound)
		stat(original, shortUrl)
	} else {
		http.NotFound(w, r)
	}
}


func stat(original, shortUrl string) {
	connstat, err := net.Dial("tcp", ":1337")
	if err != nil {
		fmt.Println("Ошибка при подключении к серверу")
		os.Exit(1)
	}
	defer connstat.Close()

	_, err = connstat.Write([]byte("1 " + original[:len(original)-1] + " http://localhost:4343/" + shortUrl + " " + getIP() + "\n"))
	if err != nil {
		fmt.Println("Ошибка при отправке команды на сервер:", err)
		return
	}
}


func getIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && ipnet.IP.To4() != nil {
			return ipnet.IP.String()
		}
	}

	return ""
}
