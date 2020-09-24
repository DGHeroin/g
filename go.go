package g

import (
    "container/list"
    "log"
    "sync/atomic"
)

type Go struct {
    maxGo          int
    runningCount   int32
    pendingGo      int32
    chanFunc       chan func()
    chanToken      chan bool

}
// waterfall context
type waterfallContext struct {
    resultCb  func(err error)
    callbacks list.List
    g         *Go
    err       error
}
// new a goroutine limiter
func New(maxGo int) *Go {
    g := &Go{
        maxGo: maxGo,
    }
    g.chanFunc = make(chan func())
    g.chanToken = make(chan bool, maxGo)
    for i := 0; i < maxGo; i++ {
        g.chanToken <- true
    }
    go func() {
        for {
            select {
            case fn := <-g.chanFunc:
                atomic.AddInt32(&g.pendingGo, -1)
                <-g.chanToken // wait for token
                g.run(fn, func() {
                    g.chanToken <- true
                })

            }
        }
    }()
    return g
}
// run a go func
func (g *Go) Go(fn func()) {
    atomic.AddInt32(&g.pendingGo, 1)
    g.chanFunc <- fn
}
// run waterfall functions
func (g *Go) Waterfall(cb func(error), callbacks ...func(err error, next func(error))) {
    ctx := &waterfallContext{
        g: g,
        resultCb: cb,
    }
    for _, f := range callbacks {
        ctx.callbacks.PushBack(f)
    }

    g.Go(func() {
        ctx.Next()
    })
}
// internal run
func (g *Go) run(fn func(), cb func()) {
    atomic.AddInt32(&g.runningCount, 1)
    go func() {
        defer func() {
            if r := recover(); r != nil {
                log.Println(r)
            }
            atomic.AddInt32(&g.runningCount, -1)
            cb()
        }()
        fn()
    }()
}
// fetch next go func to run in waterfall context
func (w *waterfallContext) Next() {
    if w.callbacks.Len() == 0 {
        w.resultCb(w.err)
        return
    }
    g := w.g
    fn := w.callbacks.Remove(w.callbacks.Front()).(func(err error, next func(error)))
    g.Go(func() {
        if w.err != nil {
            w.resultCb(w.err)
            return
        }
        fn(w.err, func(err error) {
            w.err = err
            w.Next()
        })
    })
}
