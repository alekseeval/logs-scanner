CREATE OR REPLACE FUNCTION kube_api.add_namespace(p_name varchar, p_cluster_name varchar)
RETURNS void
LANGUAGE plpgsql
AS
$$
BEGIN
    if coalesce(p_name, '') = '' then
        RAISE SQLSTATE '80001' USING message = 'empty namespace provided';
    end if;
    if coalesce(p_cluster_name, '') = '' then
        RAISE SQLSTATE '80002' USING message = 'empty cluster_name parameter provided';
    end if;
    if not EXISTS(select id from kube.clusters where name=p_cluster_name) then
        RAISE SQLSTATE '80003' USING message = 'no such cluster';
    end if;

    INSERT INTO kube.namespaces(name, cluster_name)
    VALUES (p_name, p_cluster_name);
END
$$;