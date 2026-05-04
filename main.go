package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
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

func SET(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rwm.Lock()
	defer rwm.Unlock()
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}
	value := r.URL.Query().Get("value")
	if value == "" {
		http.Error(w, "value required", http.StatusBadRequest)
		return
	}
	dataMap[key] = value
	fmt.Fprintln(w, "ok")
}

func GET(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rwm.RLock()
	defer rwm.RUnlock()
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}
	value, ok := dataMap[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintln(w, "404 not found")
		return
	}
	fmt.Fprint(w, value)
}

func DELETE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rwm.Lock()
	defer rwm.Unlock()
	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "key required", http.StatusBadRequest)
		return
	}
	delete(dataMap, key)
	fmt.Fprint(w, "ok")
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

	http.HandleFunc("/set", SET)
	http.HandleFunc("/get", GET)
	http.HandleFunc("/del", DELETE)

	shutdownChan := make(chan os.Signal, 1)
	signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Server started! Port: 8080")
		if err := http.ListenAndServe(":8080", nil); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	<-shutdownChan
	fmt.Println("\nStopping server, saving data...")
	close(stopChan)
	wg.Wait()
	save(dataFile)

}
