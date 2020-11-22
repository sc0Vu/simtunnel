# simtunnel
Simple tcp tunnel for tunneling local to remote source written in go.

# usage
```GO
tun = simtunnel.NewTunnel()
go func () {
    if err := tun.ListenAndServe("443", "google.com", "443"); err != nil {
        return
    }
}()
defer tun.Close()

// do something
```