package models

import "time"

const (
	OutboxStatusPending    = "PENDING"
	OutboxStatusProcessing = "PROCESSING"
	OutboxStatusPublished  = "PUBLISHED"
)

// OutboxEvent almacena eventos de dominio de forma transaccional para garantizar entrega eventual
type OutboxEvent struct {
	ID            string     `json:"id" gorm:"column:id;primaryKey;type:varchar(64)"`
	AggregateType string     `json:"aggregate_type" gorm:"column:aggregate_type;type:varchar(50);not null"`
	AggregateID   string     `json:"aggregate_id" gorm:"column:aggregate_id;type:varchar(64);not null;index"`
	EventType     string     `json:"event_type" gorm:"column:event_type;type:varchar(100);not null"`
	RoutingKey    string     `json:"routing_key" gorm:"column:routing_key;type:varchar(100);not null"`
	Payload       string     `json:"payload" gorm:"column:payload;type:text;not null"`
	Status        string     `json:"status" gorm:"column:status;type:varchar(20);not null;index;default:'PENDING'"`
	Attempts      int        `json:"attempts" gorm:"column:attempts;not null;default:0"`
	NextRetryAt   time.Time  `json:"next_retry_at" gorm:"column:next_retry_at;not null;index"`
	LastError     string     `json:"last_error" gorm:"column:last_error;type:text"`
	CreatedAt     time.Time  `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	PublishedAt   *time.Time `json:"published_at,omitempty" gorm:"column:published_at"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}
