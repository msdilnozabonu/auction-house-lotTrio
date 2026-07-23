update lots
set status = 'live'
where status = 'active';

alter table lots
drop constraint lots_status_check;

alter table lots
    add constraint lots_status_check
        check (status in ('draft', 'live', 'closed', 'cancelled'));