CREATE OR REPLACE FUNCTION kube_api.delete_namespace(p_cluster_name varchar, p_namespace varchar)
RETURNS void
LANGUAGE plpgsql
AS
$$
DECLARE
    r_id int;
BEGIN
    if coalesce(p_namespace, '') = '' then
        RAISE SQLSTATE '80001' USING message = 'empty namespace provided';
    end if;
    if coalesce(p_cluster_name, '') = '' then
        RAISE SQLSTATE '80002' USING message = 'empty cluster_name parameter provided';
    end if;
    if not EXISTS(select id from kube.clusters where name=p_namespace) then
        RAISE SQLSTATE '80003' USING message = 'no such cluster';
    end if;

    DELETE FROM kube.namespaces
    WHERE name=p_namespace and cluster_name=p_cluster_name
    RETURNING id INTO r_id;

    if r_id is null then
        RAISE SQLSTATE '80004' USING message = 'no such namespace';
    end if;
END
$$;