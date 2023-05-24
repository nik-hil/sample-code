package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(r.URL.Path)
		resp, err := http.Get("https://www.google.com")
		if err != nil {
			fmt.Println(err)
			return
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Fprintf(w, "%s", body)
	})
	http.ListenAndServe("localhost:8888", nil)
}
