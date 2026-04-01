-- +goose Up
alter table if exists public.users
  add column if not exists profile_text text,
  add column if not exists avatar_url text,
  add column if not exists email_visibility text;

update public.users
set email_visibility = 'approval_required'
where email_visibility is null;

alter table if exists public.users
  alter column email_visibility set default 'approval_required',
  alter column email_visibility set not null;

-- +goose StatementBegin
do $$
begin
  if not exists (
    select 1
    from pg_constraint
    where conname = 'users_email_visibility_check'
      and conrelid = 'public.users'::regclass
  ) then
    alter table public.users
      add constraint users_email_visibility_check
      check (email_visibility in ('public', 'approval_required', 'private'));
  end if;
end $$;
-- +goose StatementEnd

alter table if exists public.users
  drop column if exists proxy_email;

create table if not exists public.email_access_requests (
  id uuid primary key default gen_random_uuid(),
  requester_user_id uuid not null references public.users(id) on delete cascade,
  target_user_id uuid not null references public.users(id) on delete cascade,
  status text not null default 'pending',
  message text,
  created_at timestamptz not null default now(),
  responded_at timestamptz,
  unique (requester_user_id, target_user_id),
  constraint email_access_requests_status_check
    check (status in ('pending', 'approved', 'rejected')),
  constraint email_access_requests_not_self
    check (requester_user_id <> target_user_id)
);

-- +goose Down
drop table if exists public.email_access_requests;

alter table if exists public.users
  add column if not exists proxy_email text;

update public.users
set proxy_email = username || '@merutomo.jp'
where proxy_email is null;

create unique index if not exists users_proxy_email_key on public.users (proxy_email);

alter table if exists public.users
  alter column proxy_email set not null,
  drop constraint if exists users_email_visibility_check,
  drop column if exists email_visibility,
  drop column if exists avatar_url,
  drop column if exists profile_text;
