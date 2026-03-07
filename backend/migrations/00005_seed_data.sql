-- +goose Up
-- +goose StatementBegin
INSERT INTO accounts (code, name, type, parent_id, level, is_system, is_control) VALUES
    ('100.000', 'ASET',        'asset',     NULL, 0, true, true),
    ('200.000', 'KEWAJIBAN',   'liability', NULL, 0, true, true),
    ('300.000', 'EKUITAS',     'equity',    NULL, 0, true, true),
    ('400.000', 'PENDAPATAN',  'revenue',   NULL, 0, true, true),
    ('500.000', 'BEBAN',       'expense',   NULL, 0, true, true);

-- Insert common sub-accounts under ASET
INSERT INTO accounts (code, name, type, parent_id, level, is_system, is_control) VALUES
    ('111.000', 'Kas',             'asset', (SELECT id FROM accounts WHERE code = '100.000'), 1, true,  false),
    ('112.000', 'Piutang Usaha',   'asset', (SELECT id FROM accounts WHERE code = '100.000'), 1, true,  false),
    ('113.000', 'Persediaan',      'asset', (SELECT id FROM accounts WHERE code = '100.000'), 1, false, false);

-- Insert common sub-accounts under KEWAJIBAN
INSERT INTO accounts (code, name, type, parent_id, level, is_system, is_control) VALUES
    ('211.000', 'Hutang Usaha',    'liability', (SELECT id FROM accounts WHERE code = '200.000'), 1, true,  false),
    ('212.000', 'Hutang Pajak',    'liability', (SELECT id FROM accounts WHERE code = '200.000'), 1, false, false);

-- Insert common sub-accounts under EKUITAS
INSERT INTO accounts (code, name, type, parent_id, level, is_system, is_control) VALUES
    ('311.000', 'Modal Disetor',         'equity', (SELECT id FROM accounts WHERE code = '300.000'), 1, false, false),
    ('312.000', 'Laba Ditahan',          'equity', (SELECT id FROM accounts WHERE code = '300.000'), 1, false, false),
    ('313.000', 'Saldo Awal (Ekuitas)',  'equity', (SELECT id FROM accounts WHERE code = '300.000'), 1, true,  false);

-- Insert common sub-accounts under PENDAPATAN
INSERT INTO accounts (code, name, type, parent_id, level, is_system, is_control) VALUES
    ('411.000', 'Pendapatan Usaha',   'revenue', (SELECT id FROM accounts WHERE code = '400.000'), 1, false, false);

-- Insert common sub-accounts under BEBAN
INSERT INTO accounts (code, name, type, parent_id, level, is_system, is_control) VALUES
    ('511.000', 'Beban Operasional',  'expense', (SELECT id FROM accounts WHERE code = '500.000'), 1, false, false),
    ('512.000', 'Beban Gaji',         'expense', (SELECT id FROM accounts WHERE code = '500.000'), 1, false, false);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM accounts WHERE code IN (
    '100.000', '200.000', '300.000', '400.000', '500.000',
    '111.000', '112.000', '113.000',
    '211.000', '212.000',
    '311.000', '312.000', '313.000',
    '411.000',
    '511.000', '512.000'
);
-- +goose StatementEnd
    