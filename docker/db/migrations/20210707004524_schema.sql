-- +goose Up
-- +goose StatementBegin
CREATE TABLE Users(
    id SERIAL PRIMARY KEY,
    name varchar(100) UNIQUE,
    password varchar(100),
    public boolean DEFAULT FALSE
);

CREATE TABLE Emails(
    id SERIAL PRIMARY KEY,
    userID integer REFERENCES Users(id) ON DELETE CASCADE ON UPDATE CASCADE,
    email varchar(100) UNIQUE,
    prim boolean,
    verified boolean,
    verification_code varchar(50)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE Users;
DROP TABLE Emails;
-- +goose StatementEnd
