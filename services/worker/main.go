package main

import (
    "log"
    "os"
    "time"

    "github.com/Demosprojectstests/aegis/internal/redisx"
)

func main() {
    addr := getenv("REDIS_ADDR", "redis:6379")
    log.Println("worker redis=", addr)
    for {
        val, err := redisx.BRPop(addr, "pulse:orders", 5*time.Second)
        if err != nil {
            log.Println("brpop error:", err)
            time.Sleep(2 * time.Second)
            continue
        }
        if val == "" {
            continue
        }
        log.Println("processed", val)
    }
}

func getenv(k, d string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return d
}
