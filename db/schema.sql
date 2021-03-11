CREATE TABLE news_item(
    uid serial NOT NULL,
    news_title character varying(100) NOT NULL,
    detail text NOT NULL,
    news_image BYTEA ,
    created date,
    CONSTRAINT news_item__pkey PRIMARY KEY (uid)
) WITH (OIDS = FALSE);

CREATE TABLE project(
    uid serial NOT NULL,
    project_name character varying(150) NOT NULL,
    detail text NOT NULL,
    project_images BYTEA NOT NULL,
    start_date date,
    finish_date date,
    created date,
    CONSTRAINT project_pkey PRIMARY KEY (uid)
) WITH (OIDS = FALSE);

CREATE TABLE job_application(
    uid serial NOT NULL,
    first_name text NOT NULL,
    last_name text NOT NULL,
    email text NOT NULL,
    department text NOT NULL,
    phone_number text NOT NULL,
    cv_message text NOT NULL,
    created date,
    CONSTRAINT job_application_pkey PRIMARY KEY (uid)
) WITH (OIDS = FALSE);

CREATE TABLE relation_type(
    id serial NOT NULL,
    name character varying(32) NOT NULL,
    created date,
    CONSTRAINT relation_type_pkey PRIMARY KEY (id)
) WITH (OIDS = FALSE);

CREATE TABLE relation(
    id uuid DEFAULT uuid_generate_v4(),
    name character varying(20) NOT NULL,
    type_id bigint NOT NULL,
    path_id serial NOT NULL,
    path ltree,
    created date,
    CONSTRAINT relation_pkey PRIMARY KEY (id),
    FOREIGN KEY (type_id) REFERENCES relation_type (id)
) WITH (OIDS = FALSE);