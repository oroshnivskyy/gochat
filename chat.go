package main

import (
    "chat/config"
    "fmt"
    "log"
    "net/http"
    "text/template"
)

var homeTempl = template.Must(template.ParseFiles("templates/home.html"))
var Conf, configErr = config.Conf()

func main() {
    fmt.Println(Conf.Websocket.ReadBufSize)
    if configErr != nil {
        panic(configErr)
    }
    go h.run()
    http.HandleFunc("/", serveHome)
    http.HandleFunc("/ws", serveWs)
    err := http.ListenAndServe(Conf.Server.Name+":"+Conf.Server.Port, nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func serveHome(w http.ResponseWriter, r *http.Request) {
    if r.URL.Path != "/" {
        http.Error(w, "Not found", 404)
        return
    }
    if r.Method != "GET" {
        http.Error(w, "Method nod allowed", 405)
        return
    }
    w.Header().Set("Content-Type", "text/html; charset=utf-8")
    homeTempl.Execute(w, r.Host)
}
