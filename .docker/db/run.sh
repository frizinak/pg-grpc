#! /bin/bash
set -e

bins="/usr/lib/postgresql/16/bin"
upfile="$HOME/pgup"

echo 0 > "$upfile"

init=0
pid=
bye () {
    echo 0 > "$upfile"
    if [ "$pid" != "" ]; then
        kill $pid
    fi
}

trap bye INT QUIT TERM

if [ ! -f PG_VERSION ]; then
    echo "------------ DB INIT"
    init=1
    "$bins/pg_ctl" initdb -D .
fi

cp "$HOME/postgresql.conf" ./
cp "$HOME/pg_hba.conf" ./
"$bins/postgres" -D . &
pid=$!

sleep 0.5
while ! "$bins/pg_ctl" status -D . &>/dev/null; do
    echo "------------ DB WAITING"
    sleep 0.2
done

        #"$bins/createuser" -h /var/run/pg db_user && \
if [ $init -eq 1 ]; then
    echo "------------ DB INIT USER/DB"
    {
        "$bins/psql"     -h /var/run/pg -c "CREATE ROLE db_user \
            WITH PASSWORD 'db_pass' \
            NOSUPERUSER NOCREATEDB NOCREATEROLE INHERIT LOGIN NOREPLICATION NOBYPASSRLS;" postgres >/dev/null && \
        "$bins/createdb" -h /var/run/pg -O db_user app;
    } || \
    {
        rm -r /db/*
        exit 1
    }
fi

echo "------------ DB UP"
echo 1 > "$upfile"

while "$bins/pg_ctl" status -D . &>/dev/null; do
    sleep 1
done
pid=
bye
