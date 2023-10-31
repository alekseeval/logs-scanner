CREATE OR REPLACE FUNCTION kube_api.create_kubeconfig(p_name varchar, p_config_str varchar)
RETURNS kube.v_configs
LANGUAGE plpgsql
AS
$$
DECLARE
    r_kubeconfig kube.v_configs;
BEGIN
    if coalesce(p_name, '') = '' then
        RAISE SQLSTATE '80010' USING message = 'empty kubeconfig name provided';
    end if;
    if coalesce(p_config_str, '') = '' then
        RAISE SQLSTATE '80011' USING message = 'empty kubeconfig string provided';
    end if;

    INSERT INTO kube.kubeconfigs(name, config_str)
    VALUES (p_name, p_config_str);

    SELECT * from kube.v_configs
    WHERE name=p_name
    limit 1
    INTO r_kubeconfig;

    RETURN r_kubeconfig;
END
$$;