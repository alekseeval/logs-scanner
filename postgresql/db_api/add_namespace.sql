CREATE OR REPLACE FUNCTION kube_api.add_namespace(p_name varchar, p_kc_name varchar)
RETURNS void
LANGUAGE plpgsql
AS
$$
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

    INSERT INTO kube.namespaces(name, kc_name)
    VALUES (p_name, p_kc_name);
END
$$;