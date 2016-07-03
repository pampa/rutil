# intro

rutil is a small command line utility that lets you selectively dump, 
restore and query a redis database, peek into stored data and pretty 
print json contents.

# demo

![rutil](https://raw.githubusercontent.com/pampa/rutil/master/demo.gif)

# installation
```
  go get -v github.com/pampa/rutil
```

# usage
```
rutil [global options] command [command options] [arguments...]
```

# commands
```
dump      dump redis database to a file
pipe      dump a redis database to stdout in a format compatible with | redis-cli --pipe
restore   restore redis database from a file
query, q  query keys matching the pattern provided by --keys
help, h   Shows a list of commands or help for one command

```
# global options
```
--host, -s "127.0.0.1"  redis host
--auth, -a              authentication password
--port, -p "6379"       redis port
--help, -h              show help
--version, -v           print the version

```
   
# dump

```
rutil dump [command options] [arguments...]

```
## dump options
```
--keys, -k "*"  keys pattern (passed to redis 'keys' command)
--match, -m     regexp filter for key names
--invert, -v    invert match regexp
--auto, -a      make up a file name for the dump - redisYYYYMMDDHHMMSS.rdmp

```

# pipe

```
rutil pipe [command options] [arguments...]

```
## pipe options
```
--keys, -k "*"  keys pattern (passed to redis 'keys' command)
--match, -m     regexp filter for key names
--invert, -v    invert match regexp

```
  
# restore
```
rutil restore [command options] [arguments...]

```
## restore options
```
--dry-run, -r   pretend to restore
--flushdb, -f   flush the database before restoring
--delete, -d    delete key before restoring
--ignore, -g    ignore BUSYKEY restore errors

```
   
# query
```
rutil query [command options] [arguments...]

```
## query options
```
--keys, -k                           keys pattern (passed to redis 'keys' command)
--match, -m                          regexp filter for key names
--invert, -v                         invert match regexp
--delete                             delete keys
--print, -p                          print key values
--field, -f [-f option --f option]   hash fields to print (default all)
--json, -j                           attempt to parse and pretty print strings as json

```
   
