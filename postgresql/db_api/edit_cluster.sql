CREATE OR REPLACE FUNCTION kube_api.edit_cluster(p_name varchar, p_config_str varchar)
RETURNS kube.v_clusters
LANGUAGE plpgsql
AS
$$
DECLARE
    r_cluster kube.v_clusters;
BEGIN
    if coalesce(p_name, '') = '' then
        RAISE SQLSTATE '80010' USING message = 'empty cluster_name provided';
    end if;
    if coalesce(p_config_str, '') = '' then
        RAISE SQLSTATE '80011' USING message = 'empty config string provided';
    end if;
    IF NOT EXISTS (SELECT id from kube.clusters where name=p_name) then
        RAISE SQLSTATE '80012' USING message = 'no such cluster';
    end if;

    UPDATE kube.clusters
    SET config_str=p_config_str
    WHERE name=p_name;

    SELECT * from kube.v_clusters
    WHERE name=p_name
    limit 1
    INTO r_cluster;

    RETURN r_cluster;
END
$$;