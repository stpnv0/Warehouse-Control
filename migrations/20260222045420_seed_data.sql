-- +goose Up

-- (password: "password" для всех)
INSERT INTO users (id, username, password_hash, role) VALUES                                                      
    ('a0000000-0000-0000-0000-000000000001', 'admin',
        '$2a$10$dtFcGQacG3HbPlvSL.u0c.C68kKETjdXCTr5xMlMakTYt6KcVBAAi', 'admin'),
    ('b0000000-0000-0000-0000-000000000002', 'manager',
        '$2a$10$dtFcGQacG3HbPlvSL.u0c.C68kKETjdXCTr5xMlMakTYt6KcVBAAi', 'manager'),
    ('c0000000-0000-0000-0000-000000000003', 'viewer',
        '$2a$10$dtFcGQacG3HbPlvSL.u0c.C68kKETjdXCTr5xMlMakTYt6KcVBAAi', 'viewer')
ON CONFLICT (username) DO NOTHING;

-- +goose StatementBegin
DO $$
    BEGIN
        PERFORM set_config('app.current_user_id', 'a0000000-0000-0000-0000-000000000001', true);

        INSERT INTO items (id, name, sku, quantity, price, location) VALUES
            ('11111111-1111-1111-1111-111111111111', 'Laptop Dell XPS 15',  'DELL-XPS-15', 50,  1299.99, 'Warehouse A, Shelf 1'),
            ('22222222-2222-2222-2222-222222222222', 'Wireless Mouse',      'MS-WL-001',   200,   29.99, 'Warehouse A, Shelf 3'),
            ('33333333-3333-3333-3333-333333333333', 'USB-C Hub 7-in-1',    'HUB-USBC-7',  150,   49.99, 'Warehouse B, Shelf 2'),
            ('44444444-4444-4444-4444-444444444444', 'Monitor 27" 4K',      'MON-27-4K',    75,  449.99, 'Warehouse B, Shelf 5'),
            ('55555555-5555-5555-5555-555555555555', 'Mechanical Keyboard', 'KB-MECH-001', 120,   89.99, 'Warehouse A, Shelf 2')
        ON CONFLICT (sku) DO NOTHING;
    END;
$$;
-- +goose StatementEnd



-- +goose Down
-- Сначала отключаем триггер, чтобы DELETE FROM items не порождал новые audit-записи
ALTER TABLE items DISABLE TRIGGER trg_item_audit;

DELETE FROM items WHERE id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444',
    '55555555-5555-5555-5555-555555555555'
);

ALTER TABLE items ENABLE TRIGGER trg_item_audit;

DELETE FROM item_audit_log WHERE item_id IN (
    '11111111-1111-1111-1111-111111111111',
    '22222222-2222-2222-2222-222222222222',
    '33333333-3333-3333-3333-333333333333',
    '44444444-4444-4444-4444-444444444444',
    '55555555-5555-5555-5555-555555555555'
);

DELETE FROM users WHERE id IN (
    'a0000000-0000-0000-0000-000000000001',
    'b0000000-0000-0000-0000-000000000002',
    'c0000000-0000-0000-0000-000000000003'
);