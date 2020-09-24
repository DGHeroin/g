package g

import (
    "log"
    "sync/atomic"
)

type Go struct {
    maxGo        int
    runningCount int32
    pendingGo    int32
    chanFunc     chan func()
    chanToken    chan bool
}

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

func (g *Go) Go(fn func()) {
    atomic.AddInt32(&g.pendingGo, 1)
    g.chanFunc <- fn
}

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
