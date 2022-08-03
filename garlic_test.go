package onramp

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/eyedeekay/sam3"
)

func TestGarlic(t *testing.T) {
	fmt.Println("TestGarlic")
	garlic, err := NewGarlic("test", "127.0.0.1:7656", sam3.Options_Small)
	if err != nil {
		t.Error(err)
	}
	defer garlic.Close()
	listener, err := garlic.Listen()
	if err != nil {
		t.Error(err)
	}
	defer listener.Close()
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, %q", r.URL.Path)
	})
	go Serve(listener)
	Sleep(15)
	transport := http.Transport{
		Dial: garlic.Dial,
	}
	client := &http.Client{
		Transport: &transport,
	}
	resp, err := client.Get("http://" + listener.Addr().String() + "/")
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

}
func Serve(listener net.Listener) {
	if err := http.Serve(listener, nil); err != nil {
		log.Fatal(err)
	}
}

func Sleep(count int) {
	for i := 0; i < count; i++ {
		time.Sleep(time.Second)
		x := 15 - i
		log.Printf("Waiting: %d seconds\n", x)
	}
}
