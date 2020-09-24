package g

import (
    "fmt"
    "sync"
    "testing"
    "time"
)

func EchoI(g *Go, i int, wg *sync.WaitGroup)  {
    wg.Add(1)
    g.Go(func() {
        fmt.Println(i)
        time.Sleep(time.Second*time.Duration(i%4))
        wg.Done()
    })
}

func TestNew(t *testing.T) {
    g := New(3)
    wg := sync.WaitGroup{}
    for i := 0; i < 15; i++ {
        EchoI(g, i, &wg)
    }
    wg.Wait()
}
