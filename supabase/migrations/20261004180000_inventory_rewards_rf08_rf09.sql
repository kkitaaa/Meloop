-- Migración para soporte de RF-08, RF-09 y RN-12 (Inventario y Equipamiento de Recompensas)

-- 1. Agregar columna de habilitación en RECOMPENSA para control de vigencia / deshabilitación administrativa (RN-12)
ALTER TABLE RECOMPENSA
    ADD COLUMN IF NOT EXISTS habilitada BOOLEAN DEFAULT TRUE;

-- 2. Agregar fecha de desbloqueo en INVENTARIO (RF-09)
ALTER TABLE INVENTARIO
    ADD COLUMN IF NOT EXISTS fecha_desbloqueo TIMESTAMPTZ DEFAULT NOW();

-- 3. Índices para optimizar consultas de inventario por usuario y equipamiento
CREATE INDEX IF NOT EXISTS idx_inventario_usuario
    ON INVENTARIO(id_usuario);

CREATE INDEX IF NOT EXISTS idx_inventario_usuario_equipada
    ON INVENTARIO(id_usuario, equipada);
