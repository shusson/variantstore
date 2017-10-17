#!/bin/bash
set -e
while ! mysqladmin ping -hdb --silent; do
    echo "Waiting for db"
    sleep 3
done
mysql -hdb -p3306 -uroot -pa < /data/create.sql
