
create table directors (
    id uuid default uuid_generate_v4() primary key,
    first_name varchar(50) not null,
    last_name varchar(50) not null,
    country varchar(50) not null,
    birth_date date not null,
    has_oscar bool not null,
    constraint uq_directors UNIQUE (first_name, last_name)
);

create table movies (
    id uuid default uuid_generate_v4() primary key,
    name varchar(50) not null,
    description text,
    duration integer,
    rating numeric(3, 1),
    director_id uuid not null references directors(id)
);

create table genres(
    id uuid default uuid_generate_v4() primary key,
    name varchar(50) not null unique
);

create table movies_genres (
    movie_id uuid not null,
    genre_id uuid not null,
    constraint movie_id FOREIGN KEY (movie_id) references movies(id),
    constraint genre_id foreign key (genre_id) references genres(id)
);

create table users (
    id uuid default uuid_generate_v4() primary key,
    name varchar(50) not null unique,
    password text not null,
    email varchar(50) not null unique
);
