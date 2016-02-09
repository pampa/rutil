NAME:
   rutil - redis multitool

USAGE:
   rutil [global options] command [command options] [arguments...]
   
VERSION:
   0.1.0
   
COMMANDS:
   dump		dump redis database to a file
   restore	restore redis database from file
   help, h	Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --host, -s "127.0.0.1"	redis host
   --auth, -a 			authentication password
   --port, -p "6379"		redis port
   --help, -h			show help
   --version, -v		print the version
   
NAME:
   rutil dump - dump redis database to a file

USAGE:
   rutil dump [command options] [arguments...]

OPTIONS:
   --keys, -k "*"	dump keys that match the wildcard (passed to redis 'keys' command)
   --auto, -a		make up a file name for the dump - redisYYYYMMDDHHMMSS.rdmp
   
NAME:
   rutil restore - restore redis database from file

USAGE:
   rutil restore [command options] [arguments...]

OPTIONS:
   --dry-run, -r	pretend to restore
   --flushdb, -f	flush the database before restoring
   --delete, -d		delete key before restoring
   
