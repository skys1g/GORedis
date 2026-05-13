package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/vmihailenco/msgpack/v5"
)

var (
	dataMap = make(map[string]any)
	rwm     = sync.RWMutex{}
	wg      = sync.WaitGroup{}
)

func save(file *os.File) {
	rwm.RLock()
	bytes, err := msgpack.Marshal(dataMap)
	rwm.RUnlock()
	if err != nil {
		log.Println(err)
	}
	tmp := file.Name() + ".tmp"
	if err = os.WriteFile(tmp, bytes, 0644); err != nil {
		log.Println(err)
		return
	}
	if err = os.Rename(tmp, file.Name()); err != nil {
		log.Println(err)
		return
	}
}

func load(file *os.File) error {
	fileRead, err := os.ReadFile(file.Name())
	if err != nil {
		return err
	}
	if len(fileRead) == 0 {
		return nil
	}
	return msgpack.Unmarshal(fileRead, &dataMap)
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 3)
		cmd := strings.ToUpper(parts[0])

		switch cmd {
		case "SET":
			if len(parts) < 3 {
				fmt.Fprintln(conn, "ERR usage: SET key value")
				continue
			}
			rwm.Lock()
			dataMap[parts[1]] = parts[2]
			rwm.Unlock()
			fmt.Fprintln(conn, "OK")
		case "GET":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR usage: GET key")
				continue
			}
			rwm.RLock()
			value, ok := dataMap[parts[1]]
			rwm.RUnlock()
			if !ok {
				fmt.Fprintln(conn, "nil")
			} else {
				fmt.Fprintln(conn, value)
			}
		case "DEL":
			if len(parts) < 2 {
				fmt.Fprintln(conn, "ERR usage: DEL key")
				continue
			}
			rwm.Lock()
			delete(dataMap, parts[1])
			rwm.Unlock()
			fmt.Fprintln(conn, "OK")
		default:
			fmt.Fprintln(conn, "ERR unknown command")
		}
	}
}
func main() {
	dataFile, err := os.OpenFile("data.msgpack", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer dataFile.Close()
	err = load(dataFile)
	if err != nil {
		log.Println(err)
		return
	}
	if dataMap == nil {
		dataMap = make(map[string]any)
	}

	stopChan := make(chan struct{})
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				save(dataFile)
			case <-stopChan:
				return
			}
		}
	}()

	listener, err := net.Listen("tcp", ":6379")
	if err != nil {
		log.Println(err)
		return
	}
	go func() {
		fmt.Println("Server started! Port: 6379")
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf("Server error: %v", err)
			}
			go handleConn(conn)
		}
	}()
	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)
	<-shutdownChan
	fmt.Println("\nStopping server, saving data...")
	close(stopChan)
	wg.Wait()
	save(dataFile)

}
