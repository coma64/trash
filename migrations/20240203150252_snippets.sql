-- +goose Up
-- +goose StatementBegin
create table snippets(
    id binary(16) primary key,
    cloned_from_id binary(16),
    title text not null,
    content text not null,
    created_at timestamp not null default CURRENT_TIMESTAMP,
    foreign key (cloned_from_id) references snippets(id) on delete set null on update cascade
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
drop table snippets;
-- +goose StatementEnd
