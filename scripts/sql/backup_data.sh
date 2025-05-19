#!/bin/bash

CURR_DIR=$(
    cd $(dirname $BASH_SOURCE)
    pwd
)

# db name
DATABASE="lidb"

DB_URI="mongodb://192.168.200.60:27017"

# output directory
OUTPUT_DIR_BASE="$CURR_DIR/db_backup"
OUTPUT_DIR="$OUTPUT_DIR_BASE/$(date +%Y-%m-%d_%H-%M-%S)"
mkdir -p $OUTPUT_DIR

# collection names ; without "tasks" "logs" , if you need add them by yourself
COLLECTIONS=("groups" "jobs" "plugins" "testflows" )

# export all collection 
for COLLECTION in "${COLLECTIONS[@]}"; do
    echo "Exporting collection: $COLLECTION"
    mongodump --db $DATABASE --collection $COLLECTION --uri $DB_URI --out $OUTPUT_DIR
done

echo "All collections exported to: $OUTPUT_DIR"

# delete 15 days old backup
find $OUTPUT_DIR_BASE -type d -name "db_backup*" -mtime +15 | xargs rm -rf 


# mongorestore to restore
# mongorestore --db $DATABASE --collection $COLLECTION --uri $DB_URI --drop $OUTPUT_DIR/$COLLECTION.bson