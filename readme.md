[![GoDoc](https://godoc.org/github.com/conictus/wfe?status.svg)](https://godoc.org/github.com/conictus/wfe)
![Codeship build](https://codeship.com/projects/4f84e080-f0f0-0133-b138-3e251e5cf642/status?branch=master)

# Introduction
WFE is an asynchronous task queue/job queue based on distributed message passing. It is focused on real-time operation.
The execution units, called tasks, are executed concurrently on a single or more worker servers using multiprocessing go routines.

Tasks can execute asynchronously (in the background) or synchronously (wait until ready).
It's completely written in `GO` but totally inspired by Celery.

The main use case of this project is to build complex workflows and jobs that can run concurrently on a distributed environment easily. Note
that the project is still under development and still missing lots of features.

## Implemented features
* Asynchronous tasks execution
* Wait for a task to finish
* Tasks grouping (run multiple tasks in parallel and treat them as one)
* Tasks chaining. A chain of tasks are executed in sequence where a task result is fed as an argument to the following tasks)
* Tasks chord, which is similar to tasks group, but the results of the parallel tasks is collected and fed to a callback when all tasks are done

## Missing features
* Middleware support
* Tasks logging, needed to generate tasks graph or trees for monitoring
* Tasks routing, to make a task run on a specific worker

# How to use

## Create your work functions
First of all you need to declare your work function, the function that can be executed asynchronously on remote workers.
```go
package functions

//Note that all work function must accept an instace of `*wfe.Context` as first argument
func Add(c *wfe.Context, a, b int) int {
    return a + b
}

func Multiply(c *wfe.Context, args ...int) int {
    if len(args) == 0 {
        return 0
    }
    v := args[0]
    for i := 1; i < len(args); i++ {
        v = v * args[i]
    }

    return v
}

func init() {
    //Register the work function
    wfe.Register(Add)
    wfe.Register(Mulitply)
}
```
## Build your worker app
A worker app must import your work functions so the work functions are registered in the worker process context
```go
import (
    _ "work/functions"
)

func main() {
    //currently we only support amqp as a broker and redis as a result store
    engine, err := wfe.New(&wfe.Options{
        Broker: "amqp://localhost:5672",
        Store:  "redis://localhost:6379?keep=30",
    }, 1000)

    if err != nil {
        panic(err)
    }

    engine.Run()
}
```
## calling your tasks
A client app must import your work functions so the work function are registered in the client process context.
```go
import (
    "work/functions"
)

func main() {
    client, err := wfe.NewClient(&wfe.Options{
        Broker: "amqp://localhost:5672",
        Store:  "redis://localhost:6379?timeout=30",
    })

    if err != nil {
        panic(err)
    }

    //call functions.Add asynchronously

    task, err := client.Apply(
        wfe.MustCall(functions.Add, 10, 20),
    )

    if err != nil {
        panic(err)
    }

    log.Println(job.MustGet())

    v, e := wfe.IntResult(job.MustGet())

    //Groups, Chains and Chords

    //Group is a set of tasks that are executed in parallel.
    g, err := client.Group(
        wfe.MustCall(functions.Add, 10, 20),
        wfe.MustCall(functions.Add, 100, 200),
        wfe.MustCall(functions.Add, 1000, 200),
    )

    //reading the group result.
    for i := 0; i < g.Count(); i ++ {
        r, _ := g.ResultOf(i)
        log.Println(r.MustGet())
    }

    //Chains is a set of tasks that are executed sequentially where the first task argument is fed to the next one, and so on. until the entire chain is resolved.
    job, err := client.Chain(
        wfe.MustCall(functions.Add, 10, 20),

        wfe.MustPartialCall(functions.Add, 30),
        wfe.MustPartialCall(functions.Add, 50),
    )
    //equivalent to ((10 + 20) + 30) + 50
    log.Println(job.MustGet()) //should print 110

    //Chords are like groups with a call back
    job, err := client.Chord(
        wfe.MustPartialCall(functions.Multiply),
        wfe.MustCall(functions.Add, 10, 30),
        wfe.MustCall(functions.Add, 50, 100),
    )
    //equivalent to (10 + 30) * (50 + 100)
    log.Println(job.MustGet()) //should print 6000
}
```
