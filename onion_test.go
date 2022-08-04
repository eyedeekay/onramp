//go:build !gen
// +build !gen

package onramp

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

func TestBareOnion(t *testing.T) {
	fmt.Println("TestBareOnion Countdown")
	Sleep(5)
	onion := &Onion{}
	defer onion.Close()
	listener, err := onion.Listen()
	if err != nil {
		t.Error(err)
	}
	log.Println("listener:", listener.Addr().String())
	//defer listener.Close()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", r.URL.Path)
	})
	go Serve(listener)
	Sleep(15)
	transport := http.Transport{
		Dial: onion.Dial,
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Get("http://" + listener.Addr().String() + "/")
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Status:", resp.Status)
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("Body:", string(body))
	resp.Body.Close()

}
