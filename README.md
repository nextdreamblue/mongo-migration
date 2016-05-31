# Mongo Migration tool

These tool was built to help us migrating billions of documents between
instances of MongoDB servers.

It accepts two mongo conections.
* Specify the hosts on `--from` origin and `--to` destinity hosts
* Specify the database `--from-db`  and `--to-db` database names
* Specify the collection names from `collection-in` and `collection-out`.
  Remember the `collection-out` will be created if does not exist. Otherwise will append new data.

```
./mongo-migration --collection-in origin --collection-out origin_destination --from localhost --to localhost --from-db rdstation_development --to-db rdstation_development
```

Use `--help` to see the options
```
~/c/g/s/g/r/mongo-migration ❯❯❯ ./mongo-migration  --help
NAME:
   mongo-migration - mongo-migration --from <mongo-host-origin> --collection-in <collection-to-migrate> --to <mongo-target> --collection-out <collection-destinity>

USAGE:
   mongo-migration [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
GLOBAL OPTIONS:
   --collection-in value  collection to migrate (default: "timeline-events")
   --collection-out value collection destination name to migrate (default: "timeline-events-out")
   --from value     mongo url from origin where is the collection to migrate (default: "localhost")
   --from-db value    mongo database from origin where is the collection to migrate (default: "rdstation_development")
   --to value     mongo url destination where will be collection to migrate (default: "localhost")
   --to-db value    mongo database destination where will be the collection to migrate (default: "rdstation_development")
   --help, -h     show help
   --version, -v    print the version
```

It will only show a black screen on start. So:

* Then press `s` to `start`.
* If do you want to interrupt use `q` to do a friendly `quit`. It will finish the current batch and stop.

# How it works?

The migration strategy is only get batches of 50k records and move it to another mongo database.

If the batch fails to insert, it tries one by one and log only the records that fails individually.

It basically keeps a tracking of each record on import considering:

* If something fails while inserting: It will create a collection named `{{collection-in}}_failed` case something fails while trying to insert.
* Other records will be recorded on  `{{collection-in}}_imported` to keep the
  ids from records that was imported correctly
* Other records will be removed as soon as they are inserted in the destinity collection
* We log all stuff on the `{{collection-name}}.log`.

# Restarting the process (only test mode)

If you're testing and want to see how much it imported:

```javascript
[db.origin_imported.count(), db.origin_destination.count(),  db.origin_failed.count()] // => [ 12000000, 12000000, 0  ]
```
If do you want to restart and keep running in the same localhost for both
transference, drop the current migration collections to restart:

```javascript
[db.origin_imported.drop(), db.origin_destination.drop(),  db.origin_failed.drop()] // => [ true, true, false  ]
```

# Compiling the tool

```
go build
```

It will genereate a executable `mongo-migration`.
