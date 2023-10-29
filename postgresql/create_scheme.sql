create schema if not exists kube;
create schema if not exists kube_api;

ALTER ROLE CURRENT_ROLE SET SEARCH_PATH to kube, kube_api;

CREATE TABLE if not exists kube.kubeconfigs (
    id serial PRIMARY KEY,
    name VARCHAR(30) unique,
    config_str VARCHAR
);

CREATE TABLE if not exists kube.namespaces (
    id serial PRIMARY KEY,
    name VARCHAR,
    kc_name VARCHAR(30),

    FOREIGN KEY (kc_name) REFERENCES kube.kubeconfigs (name) ON DELETE CASCADE,
    UNIQUE (name, kc_name)
);

CREATE OR REPLACE VIEW v_configs AS
    SELECT kc.name, kc.config_str, array_agg(ns.name) filter (WHERE ns.name is not null) as namespaces
    FROM kube.kubeconfigs kc LEFT JOIN kube.namespaces ns ON kc.name = ns.kc_name
    GROUP BY kc.name, kc.config_str;
