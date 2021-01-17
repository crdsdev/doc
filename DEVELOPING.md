# Developing

## Using Postgres Docker Image

The easiest way to get started developing locally is with the official [Postgres
Docker image](https://hub.docker.com/_/postgres).

1. Start docker container in background:

```
docker run -d --rm \
   --name dev-postgres \
   -e POSTGRES_PASSWORD=password \
   -p 5432:5432 postgres
```

2. Setup doc database and tables:

```
psql -h 127.0.0.1 -U postgres -d postgres -a -f schema/crds_up.sql
```

## Using CloudSQL Proxy

If using [CloudSQL](https://cloud.google.com/sql) for a hosted Postgres
solution, the following steps can be used to develop locally against your
database.

> These steps are a summary of the
> [guide](https://cloud.google.com/sql/docs/postgres/connect-admin-proxy#docker-proxy-image)
> in the GCP CloudSQL documentation.

1. Create a `ServiceAccount` for your GCP project with the following
   permissions:
    - `Cloud SQL Admin`
    - `Cloud SQL Editor`
    - `Cloud SQL Client`
2. Click `Furnish a new private key` with type `JSON` and put it at path
   `deploy/cloudsql.json`.
3. Run the CloudSQL proxy in a docker container from this repository's root:

```
docker run -d \
  -v `pwd`/deploy:/config \
  -p 127.0.0.1:5432:5432 \
  gcr.io/cloudsql-docker/gce-proxy:1.19.1 /cloud_sql_proxy \
  -instances=crossplane-dogfood:us-central1:test-123=tcp:0.0.0.0:5432 -credential_file=/config/cloudsql.json
```

4. Setup doc database and tables:

```
psql -h 127.0.0.1 -U postgres -d postgres -a -f schema/crds_up.sql
```
