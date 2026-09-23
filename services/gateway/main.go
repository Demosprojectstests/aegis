package main

import (
    "bytes"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "time"
)

func main() {
    orders := getenv("ORDERS_URL", "http://orders")
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("ok"))
    })
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path != "/" {
            http.NotFound(w, r)
            return
        }
        fmt.Fprint(w, `{"service":"pulse-gateway"}`)
    })
    mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "POST only", http.StatusMethodNotAllowed)
            return
        }
        body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
        if len(bytes.TrimSpace(body)) == 0 {
            body = []byte(`{"item":"demo"}`)
        }
        req, err := http.NewRequest(http.MethodPost, orders+"/orders", bytes.NewReader(body))
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        req.Header.Set("Content-Type", "application/json")
        resp, err := http.DefaultClient.Do(req)
        if err != nil {
            http.Error(w, "orders unreachable: "+err.Error(), 502)
            return
        }
        defer resp.Body.Close()
        w.WriteHeader(resp.StatusCode)
        io.Copy(w, resp.Body)
    })

    s := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
    log.Println("gateway on :8080 orders=", orders)
    log.Fatal(s.ListenAndServe())
}

func getenv(k, d string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return d
}
