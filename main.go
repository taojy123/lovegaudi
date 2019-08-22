package main

import (
	"encoding/json"
	"fmt"
	"github.com/kataras/iris"
	"github.com/kataras/iris/middleware/logger"
	"github.com/kataras/iris/middleware/recover"
	bolt "go.etcd.io/bbolt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Brick struct {
	Url     string `json:"url"`
	Likes   int    `json:"likes"`
	Comment string `json:"comment"`
}

func (b Brick) Thumbnail() string {
	if !strings.HasPrefix(b.Url, "https://tse") {
		return b.Url
	}
	return b.Url + "&w=200"
}

func HandleErr(err error, title string) {
	if err != nil {
		if title == "" {
			title = "Error"
		}
		fmt.Errorf("%s: %v", title, err)
	}
}

func LoadBrick(s []byte) Brick {
	var brick Brick
	err := json.Unmarshal(s, &brick)
	HandleErr(err, "Unmarshal Error")
	//fmt.Println(brick)
	return brick
}

func DumpBrick(brick Brick) []byte {
	s, err := json.Marshal(brick)
	HandleErr(err, "Marshal Error")
	//fmt.Println(string(s))
	return s
}

func GetBricks() []Brick {
	var bricks []Brick

	db, err := bolt.Open(DB_NAME, 0666, &bolt.Options{ReadOnly: true})
	HandleErr(err, "DB Contection Error")
	defer db.Close()

	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(BUCKET_NAME)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			bricks = append(bricks, LoadBrick(v))
		}
		return nil
	})
	HandleErr(err, "")
	return bricks
}

func SaveBricks(bricks []Brick) {

	ClearBricks()

	db, err := bolt.Open(DB_NAME, 0666, nil)
	HandleErr(err, "DB Contection Error")
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BUCKET_NAME)
		err = nil
		for _, brick := range bricks {
			key := []byte(brick.Url)
			value := DumpBrick(brick)
			err := b.Put(key, value)
			HandleErr(err, "")
		}
		return err
	})
	HandleErr(err, "")
}

func ClearBricks() {
	db, err := bolt.Open(DB_NAME, 0666, nil)
	HandleErr(err, "DB Contection Error")
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(BUCKET_NAME)
		c := b.Cursor()
		err = nil
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			err := b.Delete(k)
			HandleErr(err, "")
		}
		return err
	})
	HandleErr(err, "")
}

func ShuffleBricks(bs []Brick) []Brick {
	rand.Seed(time.Now().UnixNano())
	var i, j int
	var temp Brick

	for i = len(bs) - 1; i > 0; i-- {
		j = rand.Intn(i + 1)
		temp = bs[i]
		bs[i] = bs[j]
		bs[j] = temp
	}
	return bs
}

var DB_NAME = "data.db"
var BUCKET_NAME = []byte("bricks")

func main() {

	db, err := bolt.Open(DB_NAME, 0666, nil)
	HandleErr(err, "DB Contection Error")
	db.Update(func(tx *bolt.Tx) error {
		tx.CreateBucket(BUCKET_NAME)
		return nil
	})
	db.Close()

	app := iris.New()
	app.Logger().SetLevel("debug")
	// Optionally, add two built'n handlers
	// that can recover from any http-relative panics
	// and log the requests to the terminal.
	app.Use(recover.New())
	app.Use(logger.New())

	app.RegisterView(iris.HTML("./templates", ".html"))
	app.Get("/", index)
	app.Post("/upload_brick", uploadBrick)
	app.Post("/delete_brick", deleteBrick)
	app.Get("/clear", clear)
	app.Get("/fetch", fetch)

	app.Run(iris.Addr(":8080"), iris.WithoutServerError(iris.ErrServerClosed))
}

func index(ctx iris.Context) {

	bricks := GetBricks()
	bricks = ShuffleBricks(bricks)

	ctx.ViewData("bricks", bricks)
	ctx.View("index.html")
}

func uploadBrick(ctx iris.Context) {

	bricks := GetBricks()

	url := ctx.FormValue("url")
	if !strings.HasPrefix(url, "http") {
		ctx.Redirect("/")
	}

	comment := ctx.FormValueDefault("comment", "")
	brick := Brick{Url: url, Comment: comment}
	bricks = append(bricks, brick)

	SaveBricks(bricks)

	ctx.Redirect("/")
}

func deleteBrick(ctx iris.Context) {

	url := ctx.FormValue("url")

	bricks := GetBricks()
	for i, brick := range bricks {
		if url == brick.Url {
			bricks = append(bricks[:i], bricks[i+1:]...)
			break
		}
	}
	SaveBricks(bricks)

	ctx.JSON(iris.Map{"status": "deleted"})
}

func clear(ctx iris.Context) {
	ClearBricks()
	ctx.JSON(iris.Map{"status": "clear"})
}

func fetch(ctx iris.Context) {

	bricks := GetBricks()

	//res, err := http.Get("https://www.google.com/search?q=gaudi&asearch=ichunk&async=_id:rg_s,_fmt:html")
	//HandleErr(err, "")
	//defer res.Body.Close()
	//
	//// Load the HTML document
	//doc, err := goquery.NewDocumentFromReader(res.Body)
	//fmt.Println(err)
	//
	//doc.Find("img").Each(func(i int, s *goquery.Selection) {
	//	// For each item found, get the band and title
	//	src, found := s.Attr("src")
	//	if !found {
	//		src, _ = s.Attr("data-src")
	//	}
	//	fmt.Println(src)
	//	brick := Brick{Url: src}
	//	bricks = append(bricks, brick)
	//})

	// https://cn.bing.com/images/async?q=gaudi&ensearch=1&mmasync=1&count=50&first=50

	n := ctx.URLParam("n")
	if n == "" {
		n = strconv.Itoa(rand.Intn(500))
	}
	url := "https://cn.bing.com/images/async?q=gaudi&ensearch=1&mmasync=1&count=50&first=100"

	fmt.Println(url)

	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("cache-control", "no-cache")
	res, _ := http.DefaultClient.Do(req)

	defer res.Body.Close()
	body, _ := ioutil.ReadAll(res.Body)
	content := string(body)

	//fmt.Println(res)
	fmt.Println(content)

	//exp2 := regexp.MustCompile(`src="(.*?)&`)
	exp2 := regexp.MustCompile(`&quot;(https://tse\d.mm.bing.net/th\?id=.+?)&`)
	rs := exp2.FindAllStringSubmatch(content, -1)

	for _, v := range rs {
		fmt.Println(v[0], v[1])
		brick := Brick{Url: v[1]}
		bricks = append(bricks, brick)
	}

	SaveBricks(bricks)
	ctx.Redirect("/")
}
