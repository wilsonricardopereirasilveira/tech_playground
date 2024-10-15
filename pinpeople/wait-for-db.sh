#!/bin/sh
   # wait-for-db.sh

   set -e

   host="db"
   user="pinpeople_user"
   db="pinpeople_db"
   password="PinP_s3cur3_p@ssw0rd"  # Substitua pela senha correta

   echo "Attempting to connect to database at $host:5432"
   echo "Database User: $user"
   echo "Database Name: $db"

   until PGPASSWORD=$password psql -h "$host" -U "$user" -d "$db" -c '\q'; do
     >&2 echo "Postgres is unavailable - sleeping"
     sleep 1
   done

   echo "Database is up - executing command"
   exec "$@"