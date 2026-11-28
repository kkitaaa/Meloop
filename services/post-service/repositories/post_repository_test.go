package repositories

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type createPostConnector struct {
	query string
	args  []driver.NamedValue
}

func (connector *createPostConnector) Connect(context.Context) (driver.Conn, error) {
	return &createPostConnection{connector: connector}, nil
}

func (connector *createPostConnector) Driver() driver.Driver { return connector }

func (connector *createPostConnector) Open(string) (driver.Conn, error) {
	return &createPostConnection{connector: connector}, nil
}

type createPostConnection struct {
	connector *createPostConnector
}

func (*createPostConnection) Prepare(string) (driver.Stmt, error) {
	return nil, errors.New("prepared statements are not used in this test")
}

func (*createPostConnection) Close() error { return nil }
func (*createPostConnection) Begin() (driver.Tx, error) {
	return nil, errors.New("transactions are not used in this test")
}

func (connection *createPostConnection) QueryContext(_ context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	connection.connector.query = query
	connection.connector.args = args
	return &createPostRows{id: args[0].Value.(string)}, nil
}

type createPostRows struct {
	id   string
	done bool
}

func (*createPostRows) Columns() []string {
	return []string{"id_publicacion", "id_usuario", "texto", "url_multimedia", "fecha_creacion"}
}

func (*createPostRows) Close() error { return nil }

func (rows *createPostRows) Next(values []driver.Value) error {
	if rows.done {
		return io.EOF
	}
	rows.done = true
	copy(values, []driver.Value{rows.id, "author-1", "post content", "https://cdn.example/post.png", time.Now()})
	return nil
}

func TestCreatePostInsertsAndScansCreatedPost(t *testing.T) {
	connector := &createPostConnector{}
	db := sql.OpenDB(connector)
	defer db.Close()

	mediaURL := "https://cdn.example/post.png"
	post, err := NewPostRepository(db).CreatePost(context.Background(), "author-1", "post content", &mediaURL)

	require.NoError(t, err)
	require.NotEmpty(t, post.ID)
	_, err = uuid.Parse(post.ID)
	require.NoError(t, err)
	require.Contains(t, strings.ToLower(connector.query), "insert into publicacion")
	require.Len(t, connector.args, 4)
	require.Equal(t, post.ID, connector.args[0].Value)
	require.Equal(t, "author-1", connector.args[1].Value)
	require.Equal(t, "post content", connector.args[2].Value)
	require.Equal(t, mediaURL, connector.args[3].Value)
	require.Equal(t, "author-1", post.AuthorID)
	require.Equal(t, "post content", post.Content)
	require.Equal(t, mediaURL, *post.UrlMultimedia)
}
