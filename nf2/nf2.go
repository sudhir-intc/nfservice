package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

// Config contains NF Module Configuration Data Structure
type Config struct {
	// API Root for the remote NF
	NFEndpoint     string `json:"nfendpoint"`
	LocalNfAPIRoot string `json:"localapirootprefix"`
}

type NF struct {
	Location string `json:"location"`
	Time     string `json:"time"`
}

// Path for NEF Configuration file
const cfgPath string = "config/nf.json"

var cfg Config

func main() {

	// Read the configuration
	err := loadJSONConfig(cfgPath, &cfg)
	if err != nil {
		log.Printf("Failed to load NF configuration: %v", err)
		return
	}

	// Start the Servers in a different context
	// Creating a context. This context will be used for following:
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	/* Subscribing to os Interrupt/Signal - SIGTERM and waiting for
	 * notification in a separate go routine. When the notification is received
	 * the created context will be canceled */
	osSignalCh := make(chan os.Signal, 1)
	signal.Notify(osSignalCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-osSignalCh
		log.Printf("Received signal: %#v", sig)
		cancel()
	}()
	log.Print("Starting NF2 server")
	_ = RunServer(ctx, &cfg)

}

// LoadJSONConfig reads a file located at configPath and unmarshals it to
// config structure
func loadJSONConfig(configPath string, cfg *Config) error {
	cfgData, err := ioutil.ReadFile(filepath.Clean(configPath))
	if err != nil {
		return err
	}
	err = json.Unmarshal(cfgData, cfg)
	printConfig(cfg)

	// Check if configuration is valid
	if cfg.NFEndpoint == "" {
		log.Print("NF HTTP Server endpoint  not configured")
		return errors.New("NF HTTP Server endpoint  not configured")
	}

	return err

}
func printConfig(cfg *Config) {

	log.Printf("********************* NF CONFIGURATION ******************")
	log.Printf("NF2 End Point: %v", cfg.NFEndpoint)
	log.Printf("NF2 Lcoal API Root Prefix: %v", cfg.LocalNfAPIRoot)
	log.Printf("*************************************************************")

}

func RunServer(ctx context.Context, cfg *Config) error {

	var nfserver *http.Server

	nfserver = &http.Server{
		Addr:           cfg.NFEndpoint,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	http.HandleFunc("/nf2", handlerWithCtx)

	stopServerCh := make(chan bool, 2)

	/* Go Routine is spawned here for listening for cancellation event on
	 * context */
	go func(stopServerCh chan bool) {
		<-ctx.Done()

		log.Print("Executing graceful stop for NF HTTP Server")
		if err := nfserver.Close(); err != nil {
			log.Printf("Could not close NF HTTP server: %#v", err)
		}
		log.Printf("NF HTTP server stopped")
		stopServerCh <- true
	}(stopServerCh)
	/* Go Routine is spawned here for starting NF HTTP Server */
	go startHTTPServer(nfserver, stopServerCh, "NF2")

	<-stopServerCh
	log.Print("Exiting NF2 servers")
	return nil
}

/* starting HTTP Server */
func startHTTPServer(server *http.Server,
	stopServerCh chan bool, name string) {
	if server != nil {
		log.Printf("%s HTTP 1.1 listening on %s", name, server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("HTTP server error: " + err.Error())
		}
	}
	stopServerCh <- true
}

func handlerWithCtx(w http.ResponseWriter, r *http.Request) {

	var nf1Body NF
	ctx := r.Context()

	/* Dump the request received */
	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
		return
	}
	log.Printf("NF2 Request received \n ===> %s ", string(dump))

	/* Read the response and report success if json content is proper */
	if r.Body == nil {
		log.Print("Empty Body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	// Retrieve the NF2 information from the request
	if err := json.NewDecoder(r.Body).Decode(&nf1Body); err != nil {
		log.Printf("Body parse error: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	fmt.Fprintf(w, "Hello Thanks !!!")

	defer log.Printf("NF2 Handler Completed")
	select {
	case <-time.After(1 * time.Second):
		/* Send a POST with the body received */
		client := http.Client{Timeout: 15 * time.Second}
		nf1location := nf1Body.Location

		nf1Body.Location = cfg.LocalNfAPIRoot + cfg.NFEndpoint +
			"/nf2"
		nf1Body.Time = time.Now().String()

		requestBody, err := json.Marshal(nf1Body)
		// Set request type as POST
		req, _ := http.NewRequest("POST", nf1location,
			bytes.NewBuffer(requestBody))

		// Add user-agent header and content-type header
		req.Header.Set("User-Agent", "NEF-OPENNESS-1912")
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		log.Print("Sending a request to the NF1 server")
		resp, err := client.Do(req)
		if err != nil {
			log.Print(err)
			return
		}
		defer func() {
			err = resp.Body.Close()
			if err != nil {
				log.Print("response body was not closed properly")
			}
		}()

		log.Printf("Headers in the response %d =>", resp.StatusCode)
		for k, v := range resp.Header {
			log.Printf("%q:%q\n", k, v)
		}
		log.Printf("Body in the response =>")
		respbody, err := ioutil.ReadAll(resp.Body)
		log.Print(string(respbody))

	case <-ctx.Done():
		err := ctx.Err()
		log.Print(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
