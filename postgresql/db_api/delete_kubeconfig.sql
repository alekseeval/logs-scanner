CREATE OR REPLACE FUNCTION tools_api.delete_kubeconfig(p_name varchar)
RETURNS void
LANGUAGE plpgsql
AS
$$
DECLARE
    r_kubeconfig_id int;
BEGIN
    if coalesce(p_name, '') = '' then
        RAISE SQLSTATE '80010' USING message = 'empty kubeconfig name provided';
    end if;

    DELETE FROM tools.kubeconfigs
    WHERE name=p_name
    RETURNING id INTO r_kubeconfig_id;

    if r_kubeconfig_id is null then
        RAISE SQLSTATE '80012' USING message = 'no such kubeconfig';
    end if;

END
$$;