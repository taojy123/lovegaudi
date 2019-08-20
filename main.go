package main

import (
    "fmt"
    "encoding/json"
    "github.com/kataras/iris"
    "github.com/kataras/iris/middleware/logger"
    "github.com/kataras/iris/middleware/recover"
)


type Brick struct {
    Url string `json:"url"`
    likes int `json:"likes"`
    comment string `json:"comment"`
}


func (b *Brick) Dumps() []byte{
    s, err := json.Marshal(b)    
    if err != nil {
        fmt.Errorf("Marshal Error %v", err)
    }
    fmt.Println(string(s))
    return s
}


func (b *Brick) Loads(s []byte) {
    err := json.Unmarshal(s, b)    
    if err != nil {
        fmt.Errorf("Unmarshal Error %v", err)
    }
    fmt.Println(b)
}



func main() {
    app := iris.New()
    app.Logger().SetLevel("debug")
    // Optionally, add two built'n handlers
    // that can recover from any http-relative panics
    // and log the requests to the terminal.
    app.Use(recover.New())
    app.Use(logger.New())


    // same as app.Handle("GET", "/ping", [...])
    // Method:   GET
    // Resource: http://localhost:8080/ping
    app.Get("/ping", func(ctx iris.Context) {
        ctx.WriteString("pong")
    })

    // Method:   GET
    // Resource: http://localhost:8080/hello
    app.Get("/hello", func(ctx iris.Context) {
        ctx.JSON(iris.Map{"message": "Hello Iris!"})
    })

    app.RegisterView(iris.HTML("./templates", ".html"))
    app.Get("/", func(ctx iris.Context) {

        bricks := []Brick{
            {Url: "http://via.placeholder.com/100x210"},
            {Url: "http://via.placeholder.com/100x400"},
            {Url: "http://via.placeholder.com/100x110"},
        }

        ctx.ViewData("bricks", bricks)
        ctx.View("index.html")
    })


    // http://localhost:8080
    // http://localhost:8080/ping
    // http://localhost:8080/hello
    app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}



