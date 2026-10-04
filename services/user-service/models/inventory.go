package models

import "time"

// Recompensa representa la entidad de recompensa visual en el sistema de recompensas
type Recompensa struct {
	IDRecompensa string `json:"id_recompensa"`
	Tipo         string `json:"tipo"`
	IDNivel      *int   `json:"id_nivel,omitempty"`
	Habilitada   bool   `json:"habilitada"`
}

// InventarioItem representa la relación de desbloqueo entre un usuario y una recompensa en la tabla INVENTARIO
type InventarioItem struct {
	IDInventario    string    `json:"id_inventario"`
	IDUsuario       string    `json:"id_usuario"`
	IDRecompensa    string    `json:"id_recompensa"`
	Tipo            string    `json:"tipo"`
	FechaDesbloqueo time.Time `json:"fecha_desbloqueo"`
	Equipada        bool      `json:"equipada"`
	Habilitada      bool      `json:"habilitada"`
}

// InventoryItemResponse representa la respuesta JSON estandarizada del inventario para la API (RF-09)
type InventoryItemResponse struct {
	IDInventario    string    `json:"id_inventario"`
	IDRecompensa    string    `json:"id_recompensa"`
	Tipo            string    `json:"tipo"`
	TipoRecompensa  string    `json:"tipo_recompensa,omitempty"`
	FechaDesbloqueo time.Time `json:"fecha_desbloqueo"`
	Equipada        bool      `json:"equipada"`
}
