package main

import (
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/codegangsta/cli"
	"os"
	"time"
)

var r rutil

func main() {
	app := cli.NewApp()
	app.Usage = "redis multitool utility"
	app.Version = "0.1.0"
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
				cli.BoolFlag{
					Name:  "auto, a",
					Usage: "make up a file name for the dump - redisYYYYMMDDHHMMSS.rdmp",
				},
				cli.StringFlag{
					Name:  "match, m",
					Usage: "regexp filter for key names",
				},
				cli.BoolFlag{
					Name:  "invert, v",
					Usage: "invert match regexp",
				},
			},
			Action: func(c *cli.Context) {
				args := c.Args()
				auto := c.Bool("auto")
				regex := c.String("match")
				inv := c.Bool("invert")

				var fileName string

				if len(args) == 0 && auto == false {
					checkErr("provide a file name for the dump or use rutil dump --auto to make one up")
				} else if len(args) > 0 && auto == true {
					checkErr("you can't provide a name and use --auto at the same time")
				} else if len(args) == 1 && auto == false {
					fileName = args[0]
				} else if auto == true {
					fileName = fmt.Sprintf("redis%s.rdmp", time.Now().Format("20060102150405"))
				} else if len(args) > 1 {
					checkErr("to many file names")
				} else if fileName == "" {
					checkErr("brain damage. panic")
				}

				keys, keys_c := r.getKeys(c.String("keys"), regex, inv)

				file, err := os.Create(fileName)
				checkErr(err)

				bar := pb.StartNew(keys_c)

				totalBytes := r.writeHeader(file, keys_c)

				for _, k := range keys {
					bar.Increment()
					b := r.writeDump(file, r.dumpKey(k))
					totalBytes = totalBytes + b
				}
				bar.FinishPrint(fmt.Sprintf("file: %s, keys: %d, bytes: %d", fileName, keys_c, totalBytes))
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
			},
			Action: func(c *cli.Context) {
				args := c.Args()
				dry := c.Bool("dry-run")
				flush := c.Bool("flushdb")
				del := c.Bool("delete")
				ignor := c.Bool("ignore")

				if flush && del {
					checkErr("flush or delete?")
				}

				if len(args) == 0 {
					checkErr("no file name provided")
				} else if len(args) > 1 {
					checkErr("to many file names")
				}

				file, err := os.Open(args[0])
				checkErr(err)
				hd := r.readHeader(file)

				if dry == false && flush == true {
					res := r.Client().Cmd("FLUSHDB")
					checkErr(res.Err)
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
				bar.FinishPrint(fmt.Sprintf("file: %s, keys: %d", args[0], keys_c))
			},
		},
		{
			Name:    "delete",
			Aliases: []string{"del"},
			Usage:   "delete keys matching the pattern provided by --keys",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "keys, k",
					Usage: "keys pattern (passed to redis 'keys' command)",
				},
				cli.BoolFlag{
					Name:  "yes",
					Usage: "really delete keys, default is pretend to delete",
				},
				cli.StringFlag{
					Name:  "match, m",
					Usage: "regexp filter for key names",
				},
				cli.BoolFlag{
					Name:  "invert, v",
					Usage: "invert match regexp",
				},
			},
			Action: func(c *cli.Context) {
				yes := c.Bool("yes")
				pat := c.String("keys")
				regex := c.String("match")
				inv := c.Bool("invert")
				if pat == "" {
					checkErr("missing --keys pattern")
				}

				keys, _ := r.getKeys(pat, regex, inv)

				for i, k := range keys {
					fmt.Printf("%3d: %s\n", i+1, k)
					if yes == true {
						res := r.Client().Cmd("DEL", k)
						checkErr(res.Err)
					}
				}
			},
		},
		{
			Name:    "print",
			Aliases: []string{"pp"},
			Usage:   "print keys matching the pattern provided by --keys",
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

				if pat == "" {
					checkErr("missing --keys pattern")
				}

				keys, _ := r.getKeys(pat, regex, inv)

				for _, k := range keys {
					r.printKey(k, c.StringSlice("field"), c.Bool("json"))
				}
			},
		},
	}

	app.Run(os.Args)
}
