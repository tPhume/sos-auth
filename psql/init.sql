CREATE TABLE "Users"
(
    id  serial PRIMARY KEY,
    role     VARCHAR(50)        NOT NULL,
    username     VARCHAR(50)        NOT NULL,
    email    VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(100)        NOT NULL,
    address  VARCHAR(50)
);

INSERT INTO "User" (role, username, email, password, address)
VALUES ('user', 'Jack', 'jack@supermail.com', '$2y$10$MmgAdnJGQBlI8P42Cukn4u9RpGzhtHMjg6b3Tq9Zcp2Fin/lWmMW2', 'Jack Street 42, Bangkok, Thailand');