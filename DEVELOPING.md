# Developing

# Launch Database

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

   1. Either using Docker:
   ```
   docker exec -i dev-postgres psql -U postgres < schema/crds_up.sql
   ```

   2. Or with native psql:
   ```
   psql -h 127.0.0.1 -U postgres -d postgres -a -f schema/crds_up.sql
   ```

### Using CloudSQL Proxy

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

# Start Doc server

First we need to start the worker to fetch the content. The code below assumes you deployed using a PostgreSQL Docker image.
```
docker run -d --rm \
   -p 1234:1234 \
   --link dev-postgres:pg \
   -e PG_USER=postgres \
   -e PG_PASS=password \
   -e PG_HOST=pg \
   -e PG_PORT=5432 \
   -e PG_DB=doc \
   crdsdev/doc-gitter:latest
```

Then we start the doc server properly
```
docker run -d --rm \
   -p 5000:5000 \
   --link dev-postgres:pg \
   --link doc-gitter:gitter \
   -e PG_USER=postgres \
   -e PG_PASS=password \
   -e PG_HOST=pg \
   -e PG_PORT=5432 \
   -e PG_DB=doc \
   -e GITTER_HOST=gitter \
   crdsdev/doc:latest
```
And you should be able to browse the server by hitting `http://localhost:5000`.