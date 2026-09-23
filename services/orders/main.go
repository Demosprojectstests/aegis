package main

import (
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "time"

    "github.com/Demosprojectstests/aegis/internal/redisx"
)

func main() {
    addr := getenv("REDIS_ADDR", "redis:6379")
    mux := http.NewServeMux()
    mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
        if err := redisx.Ping(addr); err != nil {
            http.Error(w, err.Error(), 503)
            return
        }
        w.Write([]byte("ok"))
    })
    mux.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "POST only", http.StatusMethodNotAllowed)
            return
        }
        body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
        id := fmt.Sprintf("ord-%d", time.Now().UnixNano())
        payload, _ := json.Marshal(map[string]string{
            "id":   id,
            "body": string(body),
        })
        if err := redisx.LPush(addr, "pulse:orders", string(payload)); err != nil {
            http.Error(w, err.Error(), 502)
            return
        }
        w.Header().Set("Content-Type", "application/json")
        fmt.Fprintf(w, `{"id":"%s","queued":true}`, id)
    })

    s := &http.Server{Addr: ":8080", Handler: mux, ReadHeaderTimeout: 5 * time.Second}
    log.Println("orders on :8080 redis=", addr)
    log.Fatal(s.ListenAndServe())
}

func getenv(k, d string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return d
}
