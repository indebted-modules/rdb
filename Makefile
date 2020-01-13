include github.com/InDebted/make/index

PG_SERVICE?=postgres
PG_USER?=indebted
PG_PWD?=indebted
DB_SCHEMA=schema/sample.sql

include github.com/InDebted/make/pg
