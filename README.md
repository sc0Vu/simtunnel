# simtunnel
Simple tcp tunnel for tunneling local to remote source written in go.

# usage

```GO
// Create a TCP tunnel, tunnel local :443 to google :443
tun = simtunnel.NewTunnel(10*time.Millisecond, 1024)
go func () {
    if err := tun.ListenAndServe("443", "google.com", "443"); err != nil {
        return
    }
}()
defer tun.Close()

// do something
```

# todo
- [ ] UDP tunnel
- [ ] Listen to 0.0.0.0
- [ ] IPV6
