package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"gopkg.in/alecthomas/kingpin.v2"
	"net/http"
	"os"
)

var (
	app = kingpin.New("skipp3r", "An S3 backed configuration utility")

	verbose = app.Flag("verbose", "Enable verbose mode").Bool()
	prefix  = app.Flag("prefix", "The base name of the config").String()
	version = app.Flag("version", "The version of the config").Default("1").String()
	bucket  = app.Flag("bucket", "The name of the config bucket").Default("ion-dev-config").String()

	get     = app.Command("get", "Get config")
	getPath = get.Arg("path", "The path").String()

	set      = app.Command("set", "Set config")
	doDelete = set.Flag("doDelete", "Use --delete with the aws cli").Bool()
	srcPath  = set.Arg("srcPath", "The source path").String()
	destPath = set.Arg("destPath", "The destination path").String()

	daemon = app.Command("daemon", "Set to run in daemon")
	port   = daemon.Arg("port", "Port to run the daemon service on").Default("16666").String()

	skipp3r Skipp3r
)

func getHandler(c *echo.Context) error {
	path := c.Param("path")
	res := skipp3r.get(&path)
	return c.String(http.StatusOK, res+"\n")
}

func main() {
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	aws_config := aws.NewConfig()

	skipp3r = Skipp3r{
		Bucket:  bucket,
		Prefix:  prefix,
		Version: version,
		Svc:     s3.New(aws_config),
	}

	switch cmd {
	case get.FullCommand():
		res := skipp3r.get(getPath)
		fmt.Println(res)
	case set.FullCommand():
		err := skipp3r.set(srcPath, destPath, doDelete)
		fmt.Println(err)
	case daemon.FullCommand():
		e := echo.New()

		e.Use(mw.Logger())
		e.Use(mw.Recover())

		e.Get("/", getHandler)
		e.Get("/:path", getHandler)
		e.Run(":" + *port)
	}
}
