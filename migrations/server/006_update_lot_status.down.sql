update lots
set status = 'active'
where status = 'live';

alter table lots
drop constraint if exists lots_status_check;

alter table lots
    add constraint lots_status_check
        check (status in ('draft', 'active', 'closed', 'cancelled'));