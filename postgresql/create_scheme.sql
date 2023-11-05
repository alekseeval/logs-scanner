create schema if not exists kube;
create schema if not exists kube_api;

ALTER ROLE CURRENT_ROLE SET SEARCH_PATH to kube, kube_api;

CREATE TABLE if not exists kube.clusters (
    id serial PRIMARY KEY,
    name VARCHAR(30) unique,
    config_str VARCHAR
);

CREATE TABLE if not exists kube.namespaces (
    id serial PRIMARY KEY,
    name VARCHAR,
    cluster_name VARCHAR(30),

    FOREIGN KEY (cluster_name) REFERENCES kube.clusters (name) ON DELETE CASCADE,
    UNIQUE (name, cluster_name)
);

CREATE OR REPLACE VIEW v_clusters AS
    SELECT kc.name, kc.config_str, coalesce(array_agg(ns.name) filter (WHERE ns.name is not null), ARRAY[]::text[]) as namespaces
    FROM kube.clusters kc LEFT JOIN kube.namespaces ns ON kc.name = ns.cluster_name
    GROUP BY kc.name, kc.config_str;
