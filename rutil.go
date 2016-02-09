package main

import (
  "os"
  "fmt"
  "io"
  "encoding/binary"
  "github.com/mediocregopher/radix.v2/redis"
)

type rutil struct {
  Host string
  Port int
  Auth string
  cli *redis.Client
}

type KeyDump struct {
  Key    []byte
  KeyL  uint64
  Dump   []byte
  DumpL uint64
  Pttl   int64
}

type FileHeader struct {
  Magic [4]byte
  Version uint8
  Keys    uint64
}

func (r *rutil) Client() *redis.Client {
  if r.cli == nil {
    cli, err := redis.Dial("tcp", fmt.Sprintf("%s:%d", r.Host, r.Port))
    checkErr(err)
    if r.Auth != "" {
      res := cli.Cmd("AUTH", r.Auth)
      checkErr(res.Err)
    }
    r.cli = cli
  }
  return r.cli
}

func (r *rutil) getKeys(wcard string) ([]string, int) {
  res := r.Client().Cmd("KEYS",wcard)
  checkErr(res.Err)
  l, err := res.List()
  checkErr(err)
  return l, len(l)
}

func (r *rutil) dumpKey(k string) KeyDump {
  d := KeyDump {
    Key: []byte(k),
    KeyL: uint64(len(k)),
  }
  var err interface{}
  cli := r.Client()
  res := cli.Cmd("PTTL",k)
  checkErr(res.Err)
  d.Pttl, err = res.Int64()
  if d.Pttl < 0 {
    d.Pttl = 0
  }
  checkErr(err)

  res = cli.Cmd("DUMP",k)
  checkErr(res.Err)
  d.Dump, err = res.Bytes()
  checkErr(err)
  d.DumpL = uint64(len(d.Dump))

  return d
}

func (r *rutil) writeHeader(f io.Writer, keys_c int) int {
  h := FileHeader {
    Magic:   [4]byte {0x52,0x44,0x4d,0x50},
    Version: uint8(0x01),
    Keys: uint64(keys_c),
  }

  checkErr(binary.Write(f, binary.BigEndian, h))
  return binary.Size(h)
}

func (r *rutil) readHeader(f io.Reader) FileHeader {
  var h FileHeader
  binary.Read(f, binary.BigEndian, &h)
  return h
}

func (r *rutil) writeDump(f io.Writer, d KeyDump) int {
  size := binary.Size(d.Pttl)  +
          binary.Size(d.KeyL)  +
          binary.Size(d.Key)   +
          binary.Size(d.DumpL) +
          binary.Size(d.Dump)

  checkErr(binary.Write(f, binary.BigEndian, d.Pttl))
  checkErr(binary.Write(f, binary.BigEndian, d.KeyL))
  checkErr(binary.Write(f, binary.BigEndian, d.Key))
  checkErr(binary.Write(f, binary.BigEndian, d.DumpL))
  checkErr(binary.Write(f, binary.BigEndian, d.Dump))
  return size
}

func (r *rutil) readDump(f io.Reader) KeyDump {
  var d KeyDump

  binary.Read(f, binary.BigEndian, &d.Pttl)
  binary.Read(f, binary.BigEndian, &d.KeyL)
  d.Key = make([]byte,d.KeyL)
  binary.Read(f, binary.BigEndian, &d.Key)
  binary.Read(f, binary.BigEndian, &d.DumpL)
  d.Dump = make([]byte,d.DumpL)
  binary.Read(f, binary.BigEndian, &d.Dump)

  return d
}

func (r *rutil) restoreKey(d KeyDump, del bool) {
  cli := r.Client()
  var res *redis.Resp

  if del {
    res = cli.Cmd("DEL",d.Key)
    checkErr(res.Err)
  }

  res = cli.Cmd("RESTORE", d.Key, d.Pttl ,d.Dump)
  checkErr(res.Err)
}

func checkErr(err interface{}) {
  if err != nil {
    fmt.Println("ERROR:",err)
    os.Exit(1)
  }
}

