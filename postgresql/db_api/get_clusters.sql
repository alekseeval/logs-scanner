CREATE OR REPLACE FUNCTION kube_api.get_clusters()
    RETURNS SETOF kube.v_clusters
LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY select * from kube.v_clusters;
END
$$;