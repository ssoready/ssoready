alter table organizations add column external_id varchar;
alter table organizations add unique (environment_id, external_id);
