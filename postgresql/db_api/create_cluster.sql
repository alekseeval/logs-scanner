CREATE OR REPLACE FUNCTION kube_api.create_cluster(p_name varchar, p_config_str varchar)
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
        RAISE SQLSTATE '80011' USING message = 'empty config_str string provided';
    end if;

    INSERT INTO kube.clusters(name, config_str)
    VALUES (p_name, p_config_str);

    SELECT * from kube.v_clusters
    WHERE name=p_name
    limit 1
    INTO r_cluster;

    RETURN r_cluster;
END
$$;