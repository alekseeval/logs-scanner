create schema if not exists tools;
create schema if not exists tools_api;

ALTER ROLE CURRENT_ROLE SET SEARCH_PATH to tools, tools_api;

CREATE TABLE if not exists tools.kubeconfigs (
    id serial PRIMARY KEY,
    name VARCHAR(30) unique,
    config_str VARCHAR
);

CREATE TABLE if not exists tools.namespaces (
    id serial PRIMARY KEY,
    name VARCHAR,
    kc_name VARCHAR(30),

    FOREIGN KEY (kc_name) REFERENCES tools.kubeconfigs (name),
    UNIQUE (name, kc_name)
);

CREATE OR REPLACE VIEW v_configs AS
    SELECT kc.name, kc.config_str, array_agg(ns.name) filter (WHERE ns.name is not null) as namespaces
    FROM kubeconfigs kc LEFT JOIN namespaces ns ON kc.name = ns.kc_name
    GROUP BY kc.name, kc.config_str;
