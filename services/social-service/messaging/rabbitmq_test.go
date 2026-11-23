package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"testing"

	"github.com/meloop/social-service/models"
	amqp "github.com/rabbitmq/amqp091-go"
)

type mockChannel struct {
	declaredExchange string
	declaredKind     string
	publishedKey     string
	publishedMand    bool
	publishedMsg     amqp.Publishing
	notifyReturnChan chan amqp.Return
	notifyRegistered bool
	returnOnClose    *amqp.Return
	callOrder        []string
	closeCalled      bool
	publishErr       error
	exchangeErr      error
	closeErr         error
}

func (m *mockChannel) ExchangeDeclare(name, kind string, durable, autoDelete, internal, noWait bool, args amqp.Table) error {
	m.callOrder = append(m.callOrder, "ExchangeDeclare")
	m.declaredExchange = name
	m.declaredKind = kind
	return m.exchangeErr
}

func (m *mockChannel) NotifyReturn(c chan amqp.Return) chan amqp.Return {
	m.callOrder = append(m.callOrder, "NotifyReturn")
	m.notifyRegistered = true
	m.notifyReturnChan = c
	return c
}

func (m *mockChannel) Publish(exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	m.callOrder = append(m.callOrder, "Publish")
	m.publishedKey = key
	m.publishedMand = mandatory
	m.publishedMsg = msg
	return m.publishErr
}

func (m *mockChannel) Close() error {
	m.callOrder = append(m.callOrder, "Close")
	if !m.closeCalled {
		m.closeCalled = true
		if m.returnOnClose != nil && m.notifyReturnChan != nil {
			m.notifyReturnChan <- *m.returnOnClose
			close(m.notifyReturnChan)
		} else if m.notifyReturnChan != nil {
			close(m.notifyReturnChan)
		}
	}
	return m.closeErr
}

type mockConnection struct {
	channel     *mockChannel
	channelErr  error
	closeCalled bool
}

func (m *mockConnection) Channel() (amqpChannel, error) {
	if m.channelErr != nil {
		return nil, m.channelErr
	}
	return m.channel, nil
}

func (m *mockConnection) Close() error {
	m.closeCalled = true
	return nil
}

func TestNewFriendRequestCancelledEventUsesContractAndUniqueUUID(t *testing.T) {
	request := &models.FriendRequest{
		ID: "request-123", SenderID: "sender-1", ReceiverID: "receiver-2",
		Status: models.FriendshipCanceled,
	}

	first, err := newFriendRequestCancelledEvent(request)
	if err != nil {
		t.Fatalf("newFriendRequestCancelledEvent() error = %v", err)
	}
	second, err := newFriendRequestCancelledEvent(request)
	if err != nil {
		t.Fatalf("newFriendRequestCancelledEvent() error = %v", err)
	}

	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(first.EventID) {
		t.Fatalf("event_id %q is not a UUID v4", first.EventID)
	}
	if first.EventID == second.EventID {
		t.Fatal("event_id must be unique for each event")
	}
	if first.EventType != friendRequestCancelledRoutingKey ||
		first.RequestID != request.ID ||
		first.SenderID != request.SenderID ||
		first.ReceiverID != request.ReceiverID ||
		first.Reason != "user_blocked" {
		t.Fatalf("unexpected event: %+v", first)
	}

	body, err := json.Marshal(first)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(body, &fields); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
	wantFields := []string{"event_id", "event_type", "request_id", "sender_id", "receiver_id", "reason"}
	if len(fields) != len(wantFields) {
		t.Fatalf("JSON fields = %v, want exactly %v", fields, wantFields)
	}
	for _, field := range wantFields {
		if _, ok := fields[field]; !ok {
			t.Errorf("JSON is missing %q", field)
		}
	}
}

func TestNewFriendRequestCancelledEventRejectsNilRequest(t *testing.T) {
	if _, err := newFriendRequestCancelledEvent(nil); err == nil {
		t.Fatal("expected an error for a nil request")
	}
}

func TestPublishFriendRequestCancelled_UnroutedDetected(t *testing.T) {
	ch := &mockChannel{
		returnOnClose: &amqp.Return{
			ReplyCode:  312,
			ReplyText:  "NO_ROUTE",
			Exchange:   exchange,
			RoutingKey: friendRequestCancelledRoutingKey,
		},
	}
	conn := &mockConnection{channel: ch}
	pub := newPublisherWithDialer("amqp://dummy", func(string) (amqpConnection, error) {
		return conn, nil
	})

	req := &models.FriendRequest{
		ID:         "req-1",
		SenderID:   "u1",
		ReceiverID: "u2",
	}

	err := pub.PublishFriendRequestCancelled(context.Background(), req)
	if err == nil {
		t.Fatal("se esperaba un error por mensaje sin ruta devuelto")
	}

	var unroutedErr *UnroutedEventError
	if !errors.As(err, &unroutedErr) {
		t.Fatalf("el error esperado era *UnroutedEventError, pero fue %T: %v", err, err)
	}

	if unroutedErr.RoutingKey != friendRequestCancelledRoutingKey {
		t.Errorf("routing_key = %q, se esperaba %q", unroutedErr.RoutingKey, friendRequestCancelledRoutingKey)
	}
	if unroutedErr.ReplyCode != 312 {
		t.Errorf("reply_code = %d, se esperaba 312", unroutedErr.ReplyCode)
	}
	if unroutedErr.Reason != "NO_ROUTE" {
		t.Errorf("reason = %q, se esperaba NO_ROUTE", unroutedErr.Reason)
	}

	if !ch.publishedMand {
		t.Error("se esperaba mandatory=true para friend.request.cancelled")
	}
	if !ch.closeCalled {
		t.Error("se esperaba que Close() fuese llamado")
	}
	if !conn.closeCalled {
		t.Error("se esperaba que la conexión fuese cerrada")
	}
}

func TestPublishFriendRequestCancelled_ValidPublication(t *testing.T) {
	ch := &mockChannel{returnOnClose: nil}
	conn := &mockConnection{channel: ch}
	pub := newPublisherWithDialer("amqp://dummy", func(string) (amqpConnection, error) {
		return conn, nil
	})

	req := &models.FriendRequest{
		ID:         "req-2",
		SenderID:   "u1",
		ReceiverID: "u2",
	}

	err := pub.PublishFriendRequestCancelled(context.Background(), req)
	if err != nil {
		t.Fatalf("no se esperaba error en publicación con ruta válida: %v", err)
	}

	if !ch.publishedMand {
		t.Error("se esperaba mandatory=true")
	}
	if ch.publishedKey != friendRequestCancelledRoutingKey {
		t.Errorf("published key = %q, se esperaba %q", ch.publishedKey, friendRequestCancelledRoutingKey)
	}
	if !ch.closeCalled {
		t.Error("se esperaba que Close() fuese llamado")
	}
}

func TestPublishFriendAccepted_PreservesMandatoryFalse(t *testing.T) {
	ch := &mockChannel{}
	conn := &mockConnection{channel: ch}
	pub := newPublisherWithDialer("amqp://dummy", func(string) (amqpConnection, error) {
		return conn, nil
	})

	req := &models.FriendRequest{
		ID:         "req-3",
		SenderID:   "u1",
		ReceiverID: "u2",
	}

	err := pub.PublishFriendAccepted(context.Background(), req)
	if err != nil {
		t.Fatalf("no se esperaba error: %v", err)
	}

	if ch.publishedMand {
		t.Error("se esperaba mandatory=false para friend.accepted")
	}
	if ch.publishedKey != "friend.accepted" {
		t.Errorf("published key = %q, se esperaba 'friend.accepted'", ch.publishedKey)
	}
	if ch.notifyRegistered {
		t.Error("no se debe registrar NotifyReturn para friend.accepted")
	}
}

func TestPublishFriendRequestCancelled_LifecycleNotifyBeforePublish(t *testing.T) {
	ch := &mockChannel{returnOnClose: nil}
	conn := &mockConnection{channel: ch}
	pub := newPublisherWithDialer("amqp://dummy", func(string) (amqpConnection, error) {
		return conn, nil
	})

	req := &models.FriendRequest{
		ID:         "req-4",
		SenderID:   "u1",
		ReceiverID: "u2",
	}

	if err := pub.PublishFriendRequestCancelled(context.Background(), req); err != nil {
		t.Fatalf("error inesperado: %v", err)
	}

	notifyIdx, publishIdx, closeIdx := -1, -1, -1
	for i, op := range ch.callOrder {
		if op == "NotifyReturn" && notifyIdx == -1 {
			notifyIdx = i
		}
		if op == "Publish" && publishIdx == -1 {
			publishIdx = i
		}
		if op == "Close" && closeIdx == -1 {
			closeIdx = i
		}
	}

	if notifyIdx == -1 || publishIdx == -1 || closeIdx == -1 {
		t.Fatalf("operaciones faltantes en callOrder: %v", ch.callOrder)
	}
	if notifyIdx >= publishIdx {
		t.Errorf("NotifyReturn (idx %d) debe ejecutarse ANTES de Publish (idx %d)", notifyIdx, publishIdx)
	}
	if publishIdx >= closeIdx {
		t.Errorf("Publish (idx %d) debe ejecutarse ANTES de Close (idx %d)", publishIdx, closeIdx)
	}
}

func TestRabbitMQIntegration_UnroutedAndRouted(t *testing.T) {
	rabbitMQURL := os.Getenv("RABBITMQ_URL")
	if rabbitMQURL == "" {
		rabbitMQURL = "amqp://guest:guest@localhost:5672/"
	}

	// Verificar si RabbitMQ está disponible
	testConn, err := amqp.Dial(rabbitMQURL)
	if err != nil {
		t.Skip("RabbitMQ no está disponible, omitiendo prueba de integración")
	}
	testConn.Close()

	pub := NewPublisher(rabbitMQURL)

	t.Run("FriendRequestCancelled_SinBinding_RetornaError", func(t *testing.T) {
		// Aseguramos un exchange limpio o usamos el existente
		req := &models.FriendRequest{
			ID:         "test-unbound-req",
			SenderID:   "user-a",
			ReceiverID: "user-b",
		}
		err := pub.PublishFriendRequestCancelled(context.Background(), req)
		if err == nil {
			t.Fatal("se esperaba un error de mensaje sin ruta al no haber cola enlazada a friend.request.cancelled")
		}
		var unroutedErr *UnroutedEventError
		if !errors.As(err, &unroutedErr) {
			t.Fatalf("tipo de error inesperado: %T: %v", err, err)
		}
		if unroutedErr.RoutingKey != friendRequestCancelledRoutingKey {
			t.Errorf("routing key = %q, se esperaba %q", unroutedErr.RoutingKey, friendRequestCancelledRoutingKey)
		}
	})

	t.Run("FriendRequestCancelled_ConBindingValido_PublicaExitosamente", func(t *testing.T) {
		conn, err := amqp.Dial(rabbitMQURL)
		if err != nil {
			t.Fatalf("error conectando a RabbitMQ: %v", err)
		}
		defer conn.Close()

		ch, err := conn.Channel()
		if err != nil {
			t.Fatalf("error abriendo canal: %v", err)
		}
		defer ch.Close()

		// Declarar cola exclusiva y enlazarla a la routing key
		q, err := ch.QueueDeclare("", false, true, true, false, nil)
		if err != nil {
			t.Fatalf("error declarando cola: %v", err)
		}

		if err := ch.QueueBind(q.Name, friendRequestCancelledRoutingKey, exchange, false, nil); err != nil {
			t.Fatalf("error vinculando cola: %v", err)
		}

		req := &models.FriendRequest{
			ID:         "test-bound-req",
			SenderID:   "user-a",
			ReceiverID: "user-b",
		}
		err = pub.PublishFriendRequestCancelled(context.Background(), req)
		if err != nil {
			t.Fatalf("no se esperaba error al haber cola vinculada: %v", err)
		}
	})

	t.Run("FriendAccepted_SinBinding_NoRetornaError", func(t *testing.T) {
		req := &models.FriendRequest{
			ID:         "test-accepted-req",
			SenderID:   "user-a",
			ReceiverID: "user-b",
		}
		err := pub.PublishFriendAccepted(context.Background(), req)
		if err != nil {
			t.Fatalf("friend.accepted con mandatory=false no debe retornar error sin binding: %v", err)
		}
	})
}
