alter table scim_user_group_memberships
    add unique (scim_user_id, scim_group_id);
