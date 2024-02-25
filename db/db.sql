create table genres(
    id uuid default uuid_generate_v4() primary key,
    name varchar(50)
);

create table directors (
    id uuid default uuid_generate_v4() primary key,
    first_name varchar(50) not null,
    last_name varchar(50) not null
);

create table movies (
    id uuid default uuid_generate_v4() primary key,
    name varchar(50) not null,
    description text,
    duration interval,
    rating numeric(3, 1),
    director_id uuid not null references directors(id),
    genre_id uuid not null references genres(id)
);

create table users (
    id uuid default uuid_generate_v4() primary key,
    name varchar(50) not null,
    password varchar(50) not null,
    email varchar(50) not null
);
