package flow

import (
    "container/list"
    "log"
    "sync"
    "sync/atomic"
)

type (
    flowContext struct {
        maxGo        int
        runningCount int32
        pendingGo    int32
        chanFunc     chan func()
        chanToken    chan bool
    }
    // waterfall context
    waterfallContext struct {
        resultCb  func(err error)
        callbacks list.List
        g         *flowContext
        err       error
    }
)

// new a goroutine limiter
func New(maxGo int) *flowContext {
    g := &flowContext{
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
func (g *flowContext) Go(fn func()) {
    atomic.AddInt32(&g.pendingGo, 1)
    g.chanFunc <- fn
}

// run waterfall functions
func (g *flowContext) Waterfall(cb func(error), callbacks ...func(err error, next func(error))) {
    ctx := &waterfallContext{
        g:        g,
        resultCb: cb,
    }
    for _, f := range callbacks {
        ctx.callbacks.PushBack(f)
    }

    g.Go(func() {
        ctx.Next()
    })
}

// run Parallel
func (g *flowContext) Parallel(cb func([]int, []error), callbacks ...func() error) {
    var (
        wg       sync.WaitGroup
        mu       sync.Mutex
        hasError bool
        funs     []int
        errs     []error
    )
    doInvoke := func(i int, fn func() error) {
        wg.Add(1)
        g.Go(func() {
            defer wg.Done()
            if hasError {
                return
            }
            if err := fn(); err != nil {
                mu.Lock()
                hasError = true
                funs = append(funs, i)
                errs = append(errs, err)
                mu.Unlock()
                return
            }
        })
    }
    for i, f := range callbacks {
        doInvoke(i, f)
    }
    wg.Wait()
    cb(funs, errs)
}

// run Parallel
func (g *flowContext) ParallelNoYield(cb func([]int, []error), callbacks ...func() error) {
    var (
        wg   sync.WaitGroup
        mu   sync.Mutex
        funs []int
        errs []error
    )
    doInvoke := func(i int, fn func() error) {
        wg.Add(1)
        g.Go(func() {
            defer wg.Done()
            if err := fn(); err != nil {
                mu.Lock()
                funs = append(funs, i)
                errs = append(errs, err)
                mu.Unlock()
                return
            }
        })
    }
    for i, f := range callbacks {
        doInvoke(i, f)
    }
    wg.Wait()
    cb(funs, errs)
}

// internal run
func (g *flowContext) run(fn func(), cb func()) {
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

func Go(f func() error) error {
    ch := make(chan error)
    go func() {
        ch <- f()
    }()
    return <-ch
}
