package messaging

import (
	"context"
	"errors"
	"testing"

	amqp "github.com/rabbitmq/amqp091-go"
)

type deliveryAcknowledgerFake struct {
	acked    bool
	rejected bool
	requeued bool
}

func (acknowledger *deliveryAcknowledgerFake) Ack(uint64, bool) error {
	acknowledger.acked = true
	return nil
}

func (acknowledger *deliveryAcknowledgerFake) Nack(_ uint64, _ bool, requeue bool) error {
	acknowledger.requeued = requeue
	return nil
}

func (acknowledger *deliveryAcknowledgerFake) Reject(uint64, bool) error {
	acknowledger.rejected = true
	return nil
}

type activityProcessorFake struct {
	err error
}

func (processor activityProcessorFake) ProcessActivity(context.Context, string, string, []byte) error {
	return processor.err
}

type permanentActivityError struct{}

func (permanentActivityError) Error() string   { return "invalid activity" }
func (permanentActivityError) Permanent() bool { return true }

func TestProcessDeliveryAcknowledgesSuccessfulActivity(t *testing.T) {
	acknowledger := &deliveryAcknowledgerFake{}
	delivery := amqp.Delivery{Acknowledger: acknowledger}

	if err := processDelivery(context.Background(), delivery, activityProcessorFake{}); err != nil {
		t.Fatalf("process delivery: %v", err)
	}
	if !acknowledger.acked || acknowledger.rejected || acknowledger.requeued {
		t.Fatalf("unexpected delivery disposition: %+v", acknowledger)
	}
}

func TestProcessDeliveryRejectsPermanentActivityErrors(t *testing.T) {
	acknowledger := &deliveryAcknowledgerFake{}
	delivery := amqp.Delivery{Acknowledger: acknowledger}

	if err := processDelivery(context.Background(), delivery, activityProcessorFake{err: permanentActivityError{}}); err != nil {
		t.Fatalf("process delivery: %v", err)
	}
	if !acknowledger.rejected || acknowledger.acked || acknowledger.requeued {
		t.Fatalf("unexpected delivery disposition: %+v", acknowledger)
	}
}

func TestProcessDeliveryRequeuesTransientActivityErrors(t *testing.T) {
	acknowledger := &deliveryAcknowledgerFake{}
	delivery := amqp.Delivery{Acknowledger: acknowledger}

	if err := processDelivery(context.Background(), delivery, activityProcessorFake{err: errors.New("Redis unavailable")}); err != nil {
		t.Fatalf("process delivery: %v", err)
	}
	if !acknowledger.requeued || acknowledger.acked || acknowledger.rejected {
		t.Fatalf("unexpected delivery disposition: %+v", acknowledger)
	}
}
