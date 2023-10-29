CREATE OR REPLACE FUNCTION kube_api.get_kubeconfig_by_name(p_kubeconfig_name varchar)
RETURNS kube.v_configs
LANGUAGE plpgsql
AS
$$
DECLARE
    r_kubeconfig kube.v_configs;
    v_cnt int;
BEGIN
    select * INTO r_kubeconfig from kube.v_configs where name=p_kubeconfig_name limit 1;
    GET DIAGNOSTICS v_cnt := ROW_COUNT;
    if v_cnt = 0 then
        RAISE SQLSTATE '80012' USING message = 'no such kubeconfig';
    end if;
    RETURN r_kubeconfig;
END
$$;