package g

import (
    "errors"
    "sync"
    "testing"
    "time"
)

func EchoI(t *testing.T,g *Go, i int, wg *sync.WaitGroup)  {
    wg.Add(1)
    g.Go(func() {
        t.Log(i)
        time.Sleep(time.Second*time.Duration(i%4))
        wg.Done()
    })
}

func TestNew(t *testing.T) {
    g := New(3)
    wg := sync.WaitGroup{}
    for i := 0; i < 10; i++ {
        EchoI(t, g, i, &wg)
    }
    wg.Wait()
}

func TestGo_Waterfall(t *testing.T) {
    g := New(3)
    g.Waterfall(func(err error) {
        t.Log("结果", err)
    }, func(err error, next func(error)) {
        t.Log("1")
        next(nil)
    }, func(err error, next func(error)) {
        t.Log("2")
        next(nil)
    }, func(err error, next func(error)) {
        t.Log("3")
        next(errors.New("error in 3"))
        //next(nil)
    }, func(err error, next func(error)) {
        t.Log("4")
        next(nil)
    })
}
