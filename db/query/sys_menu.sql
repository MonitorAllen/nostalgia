-- name: ListInitSysMenus :many
SELECT
    sm.id,
    sm.name,
    sm.path,
    sm.icon,
    sm.type,
    sm.parent_id
FROM
    sys_menus sm
        JOIN
    role_permissions rp ON sm.id = rp.menu_id
        JOIN
    admins a ON a.role_id = rp.role_id
WHERE
    a.role_id = $1
  AND sm.is_active = true AND sm.type In (1, 2) ORDER BY sort;
