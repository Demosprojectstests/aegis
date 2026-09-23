package redisx

import (
    "bufio"
    "fmt"
    "io"
    "net"
    "strconv"
    "strings"
    "time"
)

func LPush(addr, key, val string) error {
    _, err := cmd(addr, 5*time.Second, "LPUSH", key, val)
    return err
}

func BRPop(addr, key string, timeout time.Duration) (string, error) {
    sec := strconv.Itoa(int(timeout.Seconds()))
    if int(timeout.Seconds()) < 1 {
        sec = "1"
    }
    v, err := cmd(addr, timeout+3*time.Second, "BRPOP", key, sec)
    return v, err
}

func Ping(addr string) error {
    out, err := cmd(addr, 3*time.Second, "PING")
    if err != nil {
        return err
    }
    if !strings.Contains(out, "PONG") {
        return fmt.Errorf("redis ping: %s", out)
    }
    return nil
}

func cmd(addr string, deadline time.Duration, parts ...string) (string, error) {
    c, err := net.DialTimeout("tcp", addr, 3*time.Second)
    if err != nil {
        return "", err
    }
    defer c.Close()
    _ = c.SetDeadline(time.Now().Add(deadline))

    var b strings.Builder
    fmt.Fprintf(&b, "*%d\r\n", len(parts))
    for _, p := range parts {
        fmt.Fprintf(&b, "$%d\r\n%s\r\n", len(p), p)
    }
    if _, err := io.WriteString(c, b.String()); err != nil {
        return "", err
    }

    r := bufio.NewReader(c)
    line, err := r.ReadString('\n')
    if err != nil {
        return "", err
    }
    line = strings.TrimSpace(line)
    switch {
    case strings.HasPrefix(line, "+"):
        return line[1:], nil
    case strings.HasPrefix(line, ":"):
        return line[1:], nil
    case strings.HasPrefix(line, "-"):
        return "", fmt.Errorf(line[1:])
    case line == "*-1" || line == "$-1":
        return "", nil
    case strings.HasPrefix(line, "$"):
        body, err := r.ReadString('\n')
        if err != nil {
            return "", err
        }
        return strings.TrimSpace(body), nil
    case strings.HasPrefix(line, "*"):
        n, _ := strconv.Atoi(line[1:])
        var last string
        for i := 0; i < n; i++ {
            hdr, err := r.ReadString('\n')
            if err != nil {
                return "", err
            }
            if !strings.HasPrefix(strings.TrimSpace(hdr), "$") {
                continue
            }
            body, err := r.ReadString('\n')
            if err != nil {
                return "", err
            }
            last = strings.TrimSpace(body)
        }
        return last, nil
    default:
        return line, nil
    }
}
