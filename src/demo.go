package main

import (
  "github.com/garyburd/redigo/redis"
  "html/template"
  "log"
  "math/rand"
  "net/http"
  "os"
)
// data structure to hold the hit counts per host
type Hit struct {
  Host string
  Count int
}

func handler(w http.ResponseWriter, r *http.Request) {
  // The container "name" auto-generated by docker
  host := os.Getenv("HOSTNAME")
  // A pseudo variable to denote env specific variations
  env := os.Getenv("ENVIRONMENT")
  rotate := rand.Intn(180)

  // connect to redis. The redis db host should be reachable as "db"
  // Using "db" as network alias & default port
  c, err := redis.Dial("tcp", "db:6379")
  if err != nil {
    panic(err)
  }
  defer c.Close()

  // INCR the value corresponding to the host key
  c.Do("INCR", host)

  // Generate stats for all other hits per hosts
  var hits []Hit
  keys, _ := redis.Strings(c.Do("KEYS", "*"))
  for _, key := range keys {
    value, _ := redis.Int(c.Do("GET", key))
    hit := Hit{key, value}
    hits = append(hits, hit)
  }

  // Using an anonymous struct, only needed to pass to the template
  data := struct {
    CurrentHost, Env string
    Rotate int
    Hits []Hit
  }{
    host, env, rotate, hits,
  }

  // Template stuff, with error handling (critical for troubleshooting)
  t, err := template.ParseFiles("tmpl/demo.html")
  if err != nil {
    log.Fatal("Parsing error: ", err)
    return
  }

  // Voila
  exeErr := t.Execute(w, data)
  if exeErr != nil {
    log.Fatal("Execute error: ", exeErr)
  }
}

func main() {
  http.HandleFunc("/", handler)
  http.ListenAndServe(":8080", nil)
}
