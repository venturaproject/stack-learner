UPDATE telefonos SET estado_id = 5 WHERE estado_id = 4;
DELETE FROM estados_telefono WHERE id = 4;

UPDATE estados_telefono
SET nombre = 'Sin datos'
WHERE id = 5 AND nombre <> 'Sin datos';

SELECT setval(
    pg_get_serial_sequence('estados_telefono', 'id'),
    GREATEST(COALESCE((SELECT MAX(id) FROM estados_telefono), 1), 1),
    true
);
