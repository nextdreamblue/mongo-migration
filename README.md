# Mongo Migration tool

These tool was built to help us migrating billions of documents between
instances of MongoDB servers.

It accepts two mongo conections.
* Specify the database URI on `--from` origin and `--to` destinity
* Specify the collection names from `--in` and `--out`.

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

# Testing the process (only test mode)

Let's setup some data to test. Imagine you're in a localhost. So, let's play on
a `example` database name with a collection named `test`:

```javascript
use example
for(i=0;i < 10000;i++){db.test.insert({i: i});}
// WriteResult({ "nInserted" : 1  })
db.test.count() // => 10000
```

Now, let's transport these 10k collection to another database in another collection:

    ./mongo-migration \ 
      --from mongodb://localhost:27017/example \
      --to mongodb://localhost:27017/database_output \
      --in test --out output


Now after verify again on mongo client we should see something like:

```javascript
use database_output
db.output.count() // => 10000
use example
db.test.count() // => 0
```

The migration will generate some logs around batches being executed on `test.log` where `test` is the collection name from input.

# Compiling the tool

```
go build
```

It will genereate a executable `mongo-migration`.
