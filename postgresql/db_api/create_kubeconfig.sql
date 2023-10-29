CREATE OR REPLACE FUNCTION kube_api.create_kubeconfig(p_name varchar, p_config_str varchar)
RETURNS kube.kubeconfigs
LANGUAGE plpgsql
AS
$$
DECLARE
    r_kubeconfig kube.kubeconfigs;
BEGIN
    if coalesce(p_name, '') = '' then
        RAISE SQLSTATE '80010' USING message = 'empty kubeconfig name provided';
    end if;
    if coalesce(p_config_str, '') = '' then
        RAISE SQLSTATE '80011' USING message = 'empty kubeconfig string provided';
    end if;

    INSERT INTO kube.kubeconfigs(name, config_str)
    VALUES (p_name, p_config_str)
    RETURNING * INTO r_kubeconfig;

    RETURN r_kubeconfig;
END
$$;