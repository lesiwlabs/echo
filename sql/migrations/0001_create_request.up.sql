create table request (
    id uuid default gen_random_uuid(),
    created timestamp not null,
    url text not null,
    method text not null,
    ip inet,
    headers jsonb,
    body text
);

create index idx_request_created ON request (created);
create index idx_request_ip ON request (ip);
