NAME:
   rutil - redis multitool utility

USAGE:
   rutil [global options] command [command options] [arguments...]
   
VERSION:
   0.1.0
   
COMMANDS:
   dump		dump redis database to a file
   restore	restore redis database from a file
   query, q	query keys matching the pattern provided by --keys
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
   --keys, -k "*"	keys pattern (passed to redis 'keys' command)
   --auto, -a		make up a file name for the dump - redisYYYYMMDDHHMMSS.rdmp
   --match, -m 		regexp filter for key names
   --invert, -v		invert match regexp
   
NAME:
   rutil restore - restore redis database from a file

USAGE:
   rutil restore [command options] [arguments...]

OPTIONS:
   --dry-run, -r	pretend to restore
   --flushdb, -f	flush the database before restoring
   --delete, -d		delete key before restoring
   --ignore, -g		ignore BUSYKEY restore errors
   
NAME:
   rutil query - query keys matching the pattern provided by --keys

USAGE:
   rutil query [command options] [arguments...]

OPTIONS:
   --keys, -k 					keys pattern (passed to redis 'keys' command)
   --match, -m 					regexp filter for key names
   --invert, -v					invert match regexp
   --delete					delete keys
   --print, -p					print key values
   --field, -f [--field option --field option]	hash fields to print (default all)
   --json, -j					attempt to parse and pretty print strings as json
   
