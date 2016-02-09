package main

import (
  "os"
  "github.com/codegangsta/cli"
  "github.com/cheggaaa/pb"
  "fmt"
  "time"
)

var r rutil

func main() {
  app := cli.NewApp()
  app.Usage   = "redis multitool"
  app.Version = "0.1.0"
  app.Flags = []cli.Flag {
    cli.StringFlag{
      Name: "host, s",
      Value: "127.0.0.1",
      Usage: "redis host",
    },
    cli.StringFlag{
      Name: "auth, a",
      Usage: "authentication password",
    },
    cli.IntFlag {
      Name: "port, p",
      Value: 6379,
      Usage:"redis port",
    },
  }

  app.Before = func (c *cli.Context) error {
    r.Host = c.GlobalString("host")
    r.Port = c.GlobalInt("port")
    r.Auth = c.GlobalString("auth")
    return nil
  }

	app.Commands = []cli.Command{
    {
        Name:        "dump",
        Usage:       "dump redis database to a file",
        Flags: []cli.Flag {
          cli.StringFlag {
            Name: "keys, k",
            Value: "*",
            Usage: "dump keys that match the wildcard (passed to redis 'keys' command)",
          },
          cli.BoolFlag {
            Name: "auto, a",
            Usage: "make up a file name for the dump - redisYYYYMMDDHHMMSS.rdmp",
          },
        },
        Action: func(c *cli.Context) {
          args := c.Args()
          auto := c.Bool("auto")

          var fileName string

          if len(args) == 0 && auto == false  {
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

          keys, keys_c := r.getKeys(c.String("keys"))

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
        Name:        "restore",
        Usage:       "restore redis database from file",
        Flags: []cli.Flag {
          cli.BoolFlag {
            Name: "dry-run, r",
            Usage: "pretend to restore",
          },
          cli.BoolFlag {
            Name: "flushdb, f",
            Usage: "flush the database before restoring",
          },
          cli.BoolFlag {
            Name: "delete, d",
            Usage: "delete key before restoring",
          },
        },
        Action: func(c *cli.Context) {
          args  := c.Args()
          dry   := c.Bool("dry-run")
          flush := c.Bool("flushdb")
          del   := c.Bool("delete")

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
          i := uint64(0)
          for i = 0; i < hd.Keys; i++ {
            bar.Increment()
            d := r.readDump(file)
            if dry == false {
              if dry == false {
                r.restoreKey(d, del)
              }
            }
          }
          bar.FinishPrint(fmt.Sprintf("file: %s, keys: %d", args[0], i))
        },
    },
  }

  app.Run(os.Args)
}

