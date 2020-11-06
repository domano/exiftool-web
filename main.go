package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

func main() {

	var srv http.Server

	// Set up graceful shutdown of http server
	finished := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)

		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)

		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			log.Printf("shutting down http server: %v", err)
		}
		close(finished)
	}()

	// Just a single handler for our functionality
	srv.Handler = http.HandlerFunc(handleExifRequest)

	srv.Addr = "0.0.0.0:8080"

	log.Println("Starting Server on port 8080...")

	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Printf("server could not listen: %v", err)
	}

	// Block until we get the signal to finish or there is an error.
	<-finished
}
func handleExifRequest(w http.ResponseWriter, r *http.Request) {

	// Run exif command and get output
	cmd := exec.Command("exiftool", "-listx")

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("could not read from stdout: %s", err)))
		return
	}
	err = cmd.Start()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("exiftool -listx failed: %s", err)))
		return
	}

	err = DecodeXML(stdout, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("decoding exiftool failed: %s", err)))
		return
	}

	// Kill exif if connection is killed or wait for it to finish. Whichever comes first.
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()
	for {
		select {
		case <-r.Context().Done():
			if err := cmd.Process.Kill(); err != nil {
				log.Printf("could not kill exif process: %s", err.Error())
			}
			return
		case err := <-done:
			if err != nil {
				log.Printf("exif finished with error: %s", err.Error())
			}
			return
		}
	}
}
