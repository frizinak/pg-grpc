FROM ubuntu:24.04

RUN apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -q -y postgresql ca-certificates

RUN mkdir /db
WORKDIR /db

RUN groupadd db && useradd -m -g db db

COPY postgresql.conf /home/db/postgresql.conf
COPY pg_hba.conf     /home/db/pg_hba.conf
COPY run.sh          /home/db/run.sh

RUN chown -R db:db . && chmod -R 700 .

RUN mkdir -p /var/run/pg
RUN chown -R db:db /var/run/pg

VOLUME /db

USER db:db

EXPOSE 5432

ENTRYPOINT ["/home/db/run.sh"]

HEALTHCHECK --start-period=5s --timeout=1s --interval=0.5s CMD \
    cat /home/db/pgup | grep ^1|| exit 1
