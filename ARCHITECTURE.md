# Tech Stack

Maps        → MapLibre + OSM
Location    → Expo Location
Routing     → OSRM
Geocoding   → Photon (can be Selfhosted)
Backend     → Go/Gin
Realtime    → WebSocket
DB          → PostgreSQL
Cache       → Redis
Payment     → Midtrans/Dummy Payment
Push        → Expo Notifications
Database    → PgSQL & Redis (for save realtime tracking)


# Database Design

```sql
create table users (
  id uuid primary key default gen_random_uuid(),
  email varchar(150) unique not null,
  password varchar(25) not null,

  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp
);

create type casbin_ptype as enum('g', 'p');

create table casbin_rules (
  id uuid primary key default gen_random_uuid(),

  ptype casbin_ptype not null,
  v0 varchar(255),
  v1 varchar(255),
  v2 varchar(255),
  v3 varchar(255),
  v4 varchar(255),

  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp
);

create type vehicle_type as enum('car', 'motorcycle');

create table vehicles (
  id uuid primary key default gen_random_uuid(),

  name varchar(150) not null,
  licence_number varchar(20) not null,
  max_size int check(max_size > 0),
  type vehicle_type not null,

  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp
);

create table drivers (
  id uuid primary key default gen_random_uuid(),

  user_id uuid not null,
  vehicle_id uuid null,

  name varchar(255) not null,
  phone_number varchar(15) not null,
  address varchar(255) not null,
  profile_picture_url text null,

  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp,

  constraint fk_user
    foreign key (user_id)
    references users(id)
    on delete cascade,

  constraint fk_vehicle
    foreign key (vehicle_id)
    references vehicles(id)
    on delete set null
);

create table saved_address (
  id uuid primary key default gen_random_uuid(),

  customer_id uuid not null,

  name varchar(255) not null,
  lat_long varchar(40)[] not null,
  is_default_pickup boolean default false,

  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp,

  constraint fk_customer
    foreign key (customer_id)
    references customers(id)
    on delete cascade
);

create table customers (
  id uuid primary key default gen_random_uuid(),

  user_id uuid not null,

  name varchar(255) not null,
  phone_number varchar(15) not null,
  profile_picture_url text null,

  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp,

  constraint fk_user
    foreign key (user_id)
    references users(id)
    on delete cascade
);

create type transaction_status as enum('pending', 'on_progress', 'cancelled', 'completed');

create table transactions (
  id uuid primary key default gen_random_uuid(),

  customer_id uuid null,
  driver_id uuid null,
  vehicle_id uuid null,

  pickup_lat_long varchar(40)[] not null,
  destination_lat_long varchar(40)[] not null,
  last_lat_long varchar(40)[] not null,

  distance int check(distance > 0),
  fare_per_distance bigint,
  platform_percentage int,
  total_fare bigint,

  status transaction_status default 'pending',

  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp,

  constraint fk_customer
    foreign key (customer_id)
    references customers(id)
    on delete set null,

  constraint fk_driver
    foreign key (driver_id)
    references drivers(id)
    on delete set null,

  constraint fk_vehicle
    foreign key (vehicle_id)
    references vehicles(id)
    on delete set null
);

create type payout_status as enum('pending', 'processing', 'cancelled', 'paid', 'failed');

create table payout (
  id uuid primary key default gen_random_uuid(),

  driver_id uuid null,

  amount bigint,
  status payout_status default 'pending',
  failed_reason text null,

  created_at timestamp default current_timestamp,
  updated_at timestamp default current_timestamp,

  constraint fk_driver
    foreign key (driver_id)
    references drivers(id)
    on delete set null
);
```

# Authentication

JWT and JWKS accross service