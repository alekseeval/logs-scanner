CREATE OR REPLACE FUNCTION kube_api.get_kubeconfigs()
    RETURNS SETOF kube.v_configs
LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY select * from kube.v_configs;
END
$$;