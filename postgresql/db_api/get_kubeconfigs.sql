CREATE OR REPLACE FUNCTION tools_api.get_kubeconfigs()
    RETURNS SETOF tools.v_configs
LANGUAGE plpgsql
AS
$$
BEGIN
    RETURN QUERY select * from tools.v_configs;
END
$$;