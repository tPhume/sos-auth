CREATE TABLE "User"
(
    user_id  serial PRIMARY KEY,
    role     VARCHAR(50)        NOT NULL,
    username     VARCHAR(50)        NOT NULL,
    email    VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(100)        NOT NULL,
    address  VARCHAR(50)
);

INSERT INTO "User" (role, username, email, password, address)
VALUES ('user', 'Jack', 'jack@supermail.com', 'jack_password', 'Jack Street 42, Bangkok, Thailand');