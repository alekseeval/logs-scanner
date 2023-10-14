CREATE OR REPLACE FUNCTION tools_api.edit_kubeconfig(p_name varchar, p_config_str varchar)
RETURNS tools.kubeconfigs
LANGUAGE plpgsql
AS
$$
DECLARE
    r_kubeconfig tools.kubeconfigs;
BEGIN
    if coalesce(p_name, '') = '' then
        RAISE SQLSTATE '80010' USING message = 'empty kubeconfig name provided';
    end if;
    if coalesce(p_config_str, '') = '' then
        RAISE SQLSTATE '80011' USING message = 'empty kubeconfig string provided';
    end if;
    IF NOT EXISTS (SELECT id from tools.kubeconfigs where name=p_name) then
        RAISE SQLSTATE '80012' USING message = 'no such kubeconfig';
    end if;

    UPDATE tools.kubeconfigs
    SET config_str=p_config_str
    WHERE name=p_name
    RETURNING * INTO r_kubeconfig;

    RETURN r_kubeconfig;
END
$$;