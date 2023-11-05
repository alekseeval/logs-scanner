CREATE OR REPLACE FUNCTION kube_api.get_cluster_by_name(p_cluster_name varchar)
RETURNS kube.v_clusters
LANGUAGE plpgsql
AS
$$
DECLARE
    r_cluster kube.v_clusters;
    v_cnt int;
BEGIN
    select * INTO r_cluster from kube.v_clusters where name=p_cluster_name limit 1;
    GET DIAGNOSTICS v_cnt := ROW_COUNT;
    if v_cnt = 0 then
        RAISE SQLSTATE '80012' USING message = 'no such cluster';
    end if;
    RETURN r_cluster;
END
$$;