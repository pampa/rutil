package main

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"io"
	"os"
	"time"
)

var r rutil

func main() {
	app := cli.NewApp()
	app.Usage = "a collection of command line redis utils"
	app.Version = "0.1.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "host, s",
			Value: "127.0.0.1",
			Usage: "redis host",
		},
		cli.StringFlag{
			Name:  "auth, a",
			Usage: "authentication password",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 6379,
			Usage: "redis port",
		},
	}

	app.Before = func(c *cli.Context) error {
		r.Host = c.GlobalString("host")
		r.Port = c.GlobalInt("port")
		r.Auth = c.GlobalString("auth")
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "dump",
			Usage: "dump redis database to a file",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "keys, k",
					Value: "*",
					Usage: "keys pattern (passed to redis 'keys' command)",
				},
				cli.StringFlag{
					Name:  "match, m",
					Usage: "regexp filter for key names",
				},
				cli.BoolFlag{
					Name:  "invert, v",
					Usage: "invert match regexp",
				},
				cli.BoolFlag{
					Name:  "auto, a",
					Usage: "make up a file name for the dump - redisYYYYMMDDHHMMSS.rdmp",
				},
				cli.BoolFlag{
					Name:  "stdout, o",
					Usage: "dump to STDOUT",
				},
			},
			Action: func(c *cli.Context) {
				args := c.Args()
				auto := c.Bool("auto")
				regex := c.String("match")
				inv := c.Bool("invert")
				out := c.Bool("stdout")

				var fileName string

				if len(args) == 0 && auto == false && out == false {
					fail("provide a file name, --auto or --stdout")
				} else if len(args) > 0 && auto == true {
					fail("you can't provide a name and use --auto at the same time")
				} else if len(args) > 0 && out == true {
					fail("you can't provide a name and use --stdout at the same time")
				} else if auto == true && out == true {
					fail("you can't use --stdout  and --auto at the same time")
				} else if len(args) == 1 && auto == false {
					fileName = args[0]
				} else if auto == true {
					fileName = fmt.Sprintf("redis%s.rdmp", time.Now().Format("20060102150405"))
				} else if len(args) > 1 {
					fail("to many file names")
				} else if fileName == "" && out == false {
					fail("brain damage. panic")
				}

				keys, keys_c := r.getKeys(c.String("keys"), regex, inv)

				var file io.Writer
				var err interface{}
				if out {
					file = os.Stdout
				} else {
					file, err = os.Create(fileName)
					checkErr(err, "create " + fileName)
				}

				var bar *pb.ProgressBar
				if !out {
					bar = pb.StartNew(keys_c)
				}

				totalBytes := r.writeHeader(file, keys_c)

				for _, k := range keys {
					if !out {
						bar.Increment()
					}
          var ok, kd = r.dumpKey(k)
          if ok {
					  b := r.writeDump(file, kd)
					  totalBytes = totalBytes + b
          }
        }
				if !out {
					bar.FinishPrint(fmt.Sprintf("file: %s, keys: %d, bytes: %d", fileName, keys_c, totalBytes))
				}
			},
		},
		{
			Name:  "restore",
			Usage: "restore redis database from a file",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "dry-run, r",
					Usage: "pretend to restore",
				},
				cli.BoolFlag{
					Name:  "flushdb, f",
					Usage: "flush the database before restoring",
				},
				cli.BoolFlag{
					Name:  "delete, d",
					Usage: "delete key before restoring",
				},
				cli.BoolFlag{
					Name:  "ignore, g",
					Usage: "ignore BUSYKEY restore errors",
				},
				cli.BoolFlag{
					Name:  "stdin, i",
					Usage: "read dump from STDIN",
				},
			},
			Action: func(c *cli.Context) {
				args := c.Args()
				dry := c.Bool("dry-run")
				flush := c.Bool("flushdb")
				del := c.Bool("delete")
				ignor := c.Bool("ignore")
				stdin := c.Bool("stdin")

				if flush && del {
					fail("flush or delete?")
				}

				if len(args) == 0 && !stdin {
					fail("no file name provided")
				} else if len(args) > 0 && stdin {
					fail("can't use --stdin with filename")
				} else if len(args) > 1 {
					fail("to many file names")
				}

				var file io.Reader
				var fileName string

				var err interface{}
				if stdin {
					fileName = "STDIN"
					file = os.Stdin
				} else {
					fileName = args[0]
					file, err = os.Open(fileName)
					checkErr(err, "open r " + fileName)
				}
				hd := r.readHeader(file)

				if dry == false && flush == true {
					res := r.Client().Cmd("FLUSHDB")
					checkErr(res.Err, "FLUSHDB")
				}

				bar := pb.StartNew(int(hd.Keys))
				keys_c := 0
				for i := uint64(0); i < hd.Keys; i++ {
					bar.Increment()
					d := r.readDump(file)
					if dry == false {
						if dry == false {
							keys_c = keys_c + r.restoreKey(d, del, ignor)
						}
					}
				}
				bar.FinishPrint(fmt.Sprintf("file: %s, keys: %d", fileName, keys_c))
			},
		},
		{
			Name:    "query",
			Aliases: []string{"q"},
			Usage:   "query keys matching the pattern provided by --keys",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "keys, k",
					Usage: "keys pattern (passed to redis 'keys' command)",
				},
				cli.StringFlag{
					Name:  "match, m",
					Usage: "regexp filter for key names",
				},
				cli.BoolFlag{
					Name:  "invert, v",
					Usage: "invert match regexp",
				},
				cli.BoolFlag{
					Name:  "delete",
					Usage: "delete keys",
				},
				cli.BoolFlag{
					Name:  "print, p",
					Usage: "print key values",
				},
				cli.StringSliceFlag{
					Name:  "field, f",
					Usage: "hash fields to print (default all)",
				},
				cli.BoolFlag{
					Name:  "json, j",
					Usage: "attempt to parse and pretty print strings as json",
				},
			},
			Action: func(c *cli.Context) {
				pat := c.String("keys")
				regex := c.String("match")
				inv := c.Bool("invert")
				del := c.Bool("delete")
				prnt := c.Bool("print")
				flds := c.StringSlice("field")
				json := c.Bool("json")

				if pat == "" {
					fail("missing --keys pattern")
				}

				if del && prnt {
					fail("can't use --delete and --print together")
				}

				if (del || !prnt) && (json || len(flds) > 0) {
					fail("use --json and --field with --print")
				}

				keys, _ := r.getKeys(pat, regex, inv)

				for i, k := range keys {
					if prnt {
						r.printKey(k, flds, json)
					} else {
						fmt.Printf("%d: %s\n", i+1, k)
						if del == true {
							res := r.Client().Cmd("DEL", k)
							checkErr(res.Err, "DEL " + k)
						}
					}
				}
			},
		},
	}

	app.Run(os.Args)
}
