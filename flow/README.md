# go invoke flow pack

```go
// limit goroutine number

g := New(3)

for i := 0; i < 10; i++ {
    g.Go(func(){
        
    })
}
```

```go
// run a waterfall invoke chain
    g := New(3)
    g.Waterfall(func(err error) {
        t.Log("result", err)
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
```

```go
// Parallel
    g := New(3)
    g.Parallel(func(fns []int, errs []error) {
        t.Log("result", fns, errs)
    }, func() error {
        t.Log("1")
        return errors.New("error in 1")
    }, func() error {
        t.Log("2")
        return nil
    }, func() error {
        t.Log("3")
        return errors.New("error in 3")
        // next(nil)
    }, func() error {
        t.Log("4")
        return nil
    })
```