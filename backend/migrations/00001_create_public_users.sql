-- +goose Up
create table if not exists public.users (
  id uuid primary key references auth.users(id) on delete cascade,
  display_name text not null,
  username text not null unique,
  proxy_email text not null unique,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now(),
  constraint users_username_format check (username ~ '^[a-z0-9_]{3,30}$')
);

-- +goose StatementBegin
create or replace function public.set_updated_at()
returns trigger
language plpgsql
as $$
begin
  new.updated_at = now();
  return new;
end;
$$;
-- +goose StatementEnd

drop trigger if exists users_set_updated_at on public.users;

create trigger users_set_updated_at
before update on public.users
for each row
execute function public.set_updated_at();

-- +goose Down
drop trigger if exists users_set_updated_at on public.users;
drop function if exists public.set_updated_at();
drop table if exists public.users;
