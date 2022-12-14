//go:build !gen
// +build !gen

package onramp

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestBareGarlic(t *testing.T) {
	fmt.Println("TestBareGarlic Countdown")
	Sleep(5)
	garlic := &Garlic{}
	defer garlic.Close()
	listener, err := garlic.ListenTLS()
	if err != nil {
		t.Error(err)
	}
	log.Println("listener:", listener.Addr().String())
	defer listener.Close()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", r.URL.Path)
	})
	go Serve(listener)
	Sleep(15)
	transport := http.Transport{
		Dial: garlic.Dial,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Get("https://" + listener.Addr().String() + "/")
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	fmt.Println(resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(string(body))
	Sleep(5)
}

func Serve(listener net.Listener) {
	if err := http.Serve(listener, nil); err != nil {
		log.Fatal(err)
	}
}

func Sleep(count int) {
	for i := 0; i < count; i++ {
		time.Sleep(time.Second)
		x := count - i
		log.Printf("Waiting: %d seconds\n", x)
	}
}
