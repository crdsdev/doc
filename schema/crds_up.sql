CREATE DATABASE doc;

\connect doc;

CREATE TABLE crds (
    "group" VARCHAR(255) NOT NULL,
    version VARCHAR(255) NOT NULL,
    kind VARCHAR(255) NOT NULL,
    repo VARCHAR (255) NOT NULL,
    tag VARCHAR(255) NOT NULL,
    filename VARCHAR(255) NOT NULL,
    data JSONB NOT NULL,
    PRIMARY KEY(repo, tag, "group", version, kind)
);
