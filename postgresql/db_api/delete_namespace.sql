CREATE OR REPLACE FUNCTION kube_api.delete_namespace(p_name varchar, p_kc_name varchar)
RETURNS void
LANGUAGE plpgsql
AS
$$
DECLARE
    r_id int;
BEGIN
    if coalesce(p_name, '') = '' then
        RAISE SQLSTATE '80001' USING message = 'empty namespace name provided';
    end if;
    if coalesce(p_kc_name, '') = '' then
        RAISE SQLSTATE '80002' USING message = 'empty kubeconfig name parameter provided';
    end if;
    if not EXISTS(select id from kube.kubeconfigs where name=p_kc_name) then
        RAISE SQLSTATE '80003' USING message = 'no such kubeconfig';
    end if;

    DELETE FROM kube.namespaces
    WHERE name=p_name and kc_name=p_kc_name
    RETURNING id INTO r_id;

    if r_id is null then
        RAISE SQLSTATE '80004' USING message = 'no such namespace';
    end if;
END
$$;