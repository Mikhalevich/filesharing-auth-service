-- +goose Up
-- +goose StatementBegin
INSERT INTO Users(name, public) VALUES('common', TRUE);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM Users WHERE name = 'common';
-- +goose StatementEnd
