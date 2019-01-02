package pgconn_test

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/pgconn"
	"github.com/stretchr/testify/require"
)

func BenchmarkConnect(b *testing.B) {
	benchmarks := []struct {
		name string
		env  string
	}{
		{"Unix socket", "PGX_TEST_UNIX_SOCKET_CONN_STRING"},
		{"TCP", "PGX_TEST_TCP_CONN_STRING"},
	}

	for _, bm := range benchmarks {
		b.Run(bm.name, func(b *testing.B) {
			connString := os.Getenv(bm.env)
			if connString == "" {
				b.Skipf("Skipping due to missing environment variable %v", bm.env)
			}

			for i := 0; i < b.N; i++ {
				conn, err := pgconn.Connect(context.Background(), connString)
				require.Nil(b, err)

				err = conn.Close(context.Background())
				require.Nil(b, err)
			}
		})
	}
}

func BenchmarkExec(b *testing.B) {
	conn, err := pgconn.Connect(context.Background(), os.Getenv("PGX_TEST_DATABASE"))
	require.Nil(b, err)
	defer closeConn(b, conn)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := conn.Exec(context.Background(), "select 'hello'::text as a, 42::int4 as b, '2019-01-01'::date")
		require.Nil(b, err)
	}
}

func BenchmarkExecPrepared(b *testing.B) {
	conn, err := pgconn.Connect(context.Background(), os.Getenv("PGX_TEST_DATABASE"))
	require.Nil(b, err)
	defer closeConn(b, conn)

	_, err = conn.Prepare(context.Background(), "ps1", "select 'hello'::text as a, 42::int4 as b, '2019-01-01'::date", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := conn.ExecPrepared(context.Background(), "ps1", nil, nil, nil)
		require.Nil(b, err)
	}
}

func BenchmarkSendExecPrepared(b *testing.B) {
	conn, err := pgconn.Connect(context.Background(), os.Getenv("PGX_TEST_DATABASE"))
	require.Nil(b, err)
	defer closeConn(b, conn)

	_, err = conn.Prepare(context.Background(), "ps1", "select 'hello'::text as a, 42::int4 as b, '2019-01-01'::date", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		conn.SendExecPrepared("ps1", nil, nil, nil)
		err := conn.Flush(context.Background())
		require.Nil(b, err)

		for conn.NextResult(context.Background()) {
			_, err := conn.ResultReader().Close()
			require.Nil(b, err)
		}
	}
}
