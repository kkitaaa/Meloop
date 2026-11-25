-- RN-05: persistencia minima para calcular inactividad en amigos en comun.
-- Las cuentas existentes quedan con fecha_creacion NULL: no hay evidencia suficiente
-- para crear una fecha historica ni clasificarlas como inactivas.
ALTER TABLE USUARIO
    ADD COLUMN IF NOT EXISTS fecha_creacion TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS fecha_ultimo_login TIMESTAMPTZ;

