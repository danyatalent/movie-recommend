create table genres(
    id serial not null primary key,
    name varchar(50)
);

create table directors (
    id serial primary key,
    first_name varchar(50) not null,
    last_name varchar(50) not null
);

create table movies (
    id serial primary key,
    name varchar(50) not null,
    description text,
    duration interval,
    rating numeric(3, 1),
    director_id integer not null references directors(id),
    genre_id integer not null references genres(id)
);

create table users (
    id serial primary key,
    name varchar(50) not null,
    password varchar(50) not null,
    email varchar(50) not null
)