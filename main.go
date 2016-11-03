package main

import (
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

func handler1(w http.ResponseWriter, r *http.Request, counter Counter){
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "GET"{
		//rUrl := r.URL.Query().Get("url")
		//counter.count++
		w.Write([]byte(strconv.Itoa(counter.count)))
	}else{
		counter.count++
		w.Write([]byte(strconv.Itoa(counter.count)))
	}
}

type Counter struct{
	count int
}

func main(){

	counter := Counter{0}
	handler := func(w http.ResponseWriter, r *http.Request) {
		handler1(w,r,counter)
	}
	r := mux.NewRouter()
	r.HandleFunc("/", handler)
	http.Handle("/",r)
	http.ListenAndServe(":8080", nil)

}
