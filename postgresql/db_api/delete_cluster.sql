CREATE OR REPLACE FUNCTION kube_api.delete_cluster(p_name varchar)
RETURNS void
LANGUAGE plpgsql
AS
$$
DECLARE
    r_cluster_id int;
BEGIN
    if coalesce(p_name, '') = '' then
        RAISE SQLSTATE '80010' USING message = 'empty cluster_name provided';
    end if;

    DELETE FROM kube.clusters
    WHERE name=p_name
    RETURNING id INTO r_cluster_id;

    if r_cluster_id is null then
        RAISE SQLSTATE '80012' USING message = 'no such cluster';
    end if;

END
$$;