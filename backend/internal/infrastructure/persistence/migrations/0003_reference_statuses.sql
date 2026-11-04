INSERT INTO estados_telefono (id, nombre, color) VALUES
    (1, 'Activo', 'teal'),
    (2, 'Congelado', 'sky'),
    (3, 'Baja', 'gray'),
    (5, 'Sin datos', 'amber')
ON CONFLICT (id) DO NOTHING;

SELECT setval(
    pg_get_serial_sequence('estados_telefono', 'id'),
    GREATEST(COALESCE((SELECT MAX(id) FROM estados_telefono), 1), 1),
    true
);
