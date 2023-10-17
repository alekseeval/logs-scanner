CREATE OR REPLACE FUNCTION tools_api.get_kubeconfig_by_name(p_kubeconfig_name varchar)
RETURNS tools.v_configs
LANGUAGE plpgsql
AS
$$
DECLARE
    r_kubeconfig tools.v_configs;
    v_cnt int;
BEGIN
    select * INTO r_kubeconfig from tools.v_configs where name=p_kubeconfig_name limit 1;
    GET DIAGNOSTICS v_cnt := ROW_COUNT;
    if v_cnt = 0 then
        RAISE SQLSTATE '80012' USING message = 'no such kubeconfig';
    end if;
    RETURN r_kubeconfig;
END
$$;