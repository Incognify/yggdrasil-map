package main

import (
  "encoding/json"
  "fmt"
  "net"
  "sync"
  "time"
)

var waitgroup sync.WaitGroup
var visited sync.Map
var rumored sync.Map

func dial() (net.Conn, error) {
  return net.DialTimeout("unix", "/var/run/yggdrasil.sock", time.Second)
}

func getRequest(key, request string) map[string]interface{} {
  return map[string]interface{}{
    "keepalive": true,
    "request": request,
    "key": key,
  }
}

func doRequest(request map[string]interface{}) map[string]interface{} {
  req, err := json.Marshal(request)
  if err != nil {
    panic(err)
  }
  sock, err := dial()
  if err != nil {
    panic(err)
  }
  if _,err = sock.Write(req); err != nil {
    panic(err)
  }
  bs := make([]byte, 65535)
  n, err := sock.Read(bs)
  if err != nil {
    panic(bs)
  }
  bs = bs[:n]
  var res map[string]interface{}
  if err = json.Unmarshal(bs, &res); err != nil {
    panic(err)
  }
  return res
}

func getNodeInfo(key string) map[string]interface{} {
  return doRequest(getRequest(key, "getNodeInfo"))
}

func getSelf(key string) map[string]interface{} {
  return doRequest(getRequest(key, "debug_remoteGetSelf"))
}

func getPeers(key string) map[string]interface{} {
  return doRequest(getRequest(key, "debug_remoteGetPeers"))
}

func getDHT(key string) map[string]interface{} {
  return doRequest(getRequest(key, "debug_remoteGetDHT"))
}

type rumorResult struct {
  key string
  res map[string]interface{}
}

func doRumor(key string, out chan rumorResult) {
  waitgroup.Add(1)
  go func() {
    defer waitgroup.Done()
    if _, known := rumored.LoadOrStore(key, true); known {
      return
    }
    defer rumored.Delete(key)
    if _, known := visited.Load(key); known {
      return
    }
    results := make(map[string]interface{})
    if res, ok := getNodeInfo(key)["response"]; ok {
      for addr,v := range res.(map[string]interface{}) {
        results["address"] = addr
        results["nodeinfo"] = v
      }
    }
    if res, ok := getSelf(key)["response"]; ok {
      for _,v := range res.(map[string]interface{}) {
        vm := v.(map[string]interface{})
        if coords, ok := vm["coords"]; ok {
          results["coords"] = coords
        }
      }
    }
    if res, ok := getPeers(key)["response"]; ok {
      for _,v := range res.(map[string]interface{}) {
        vm := v.(map[string]interface{})
        if keys, ok := vm["keys"]; ok {
          results["peers"] = keys
          for _,key := range keys.([]interface{}) {
            doRumor(key.(string), out)
          }
        }
      }
    }
    if res, ok := getDHT(key)["response"]; ok {
      for _,v := range res.(map[string]interface{}) {
        vm := v.(map[string]interface{})
        if keys, ok := vm["keys"]; ok {
          results["dht"] = keys
          for _,key := range keys.([]interface{}) {
            doRumor(key.(string), out)
          }
        }
      }
    }
    if len(results) > 0 {
      if _, known := visited.LoadOrStore(key, true); known {
        return
      }
      results["time"] = time.Now().Unix()
      out<-rumorResult{key, results}
    }
  }()
}

func doPrinter() (chan rumorResult, chan struct{}) {
  results := make(chan rumorResult)
  done := make(chan struct{})
  go func() {
    defer close(done)
    fmt.Println("{\"yggnodes\": {")
    var notFirst bool
    for result := range results {
      // TODO correct output
      res, err := json.Marshal(result.res)
      if err != nil {
        panic(err)
      }
      if notFirst {
        fmt.Println(",")
      }
      fmt.Printf("\"%s\": %s", result.key, res)
      notFirst = true
    }
    fmt.Println("\n}}")
  }()
  return results, done
}

func main() {
  self := doRequest(map[string]interface{}{"keepalive": true, "request": "getSelf"})
  res := self["response"].(map[string]interface{})["self"].(map[string]interface{})
  var key string
  for _,v := range res {
    key = v.(map[string]interface{})["key"].(string)
  }
  results, done := doPrinter()
  doRumor(key, results)
  waitgroup.Wait()
  close(results)
  <-done
}
