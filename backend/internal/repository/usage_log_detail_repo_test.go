package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestUsageLogDetailRepositoryBatchHasDetailsByRequestKeys_CastsCTEParams(t *testing.T) {
	db, mock := newSQLMock(t)
	repo := &usageLogDetailRepository{sql: db}

	keys := []service.UsageLogDetailKey{
		{RequestID: "req-1", APIKeyID: 2},
		{RequestID: "req-2", APIKeyID: 3},
	}

	mock.ExpectQuery("WITH input\\(request_id, api_key_id\\) AS \\(\\s*VALUES \\(\\$1::text, \\$2::bigint\\), \\(\\$3::text, \\$4::bigint\\)\\s*\\)\\s*SELECT d.request_id, d.api_key_id\\s*FROM usage_log_details d\\s*INNER JOIN input i\\s*ON i.request_id = d.request_id\\s*AND i.api_key_id = d.api_key_id").
		WithArgs("req-1", int64(2), "req-2", int64(3)).
		WillReturnRows(sqlmock.NewRows([]string{"request_id", "api_key_id"}).AddRow("req-1", int64(2)))

	found, err := repo.BatchHasDetailsByRequestKeys(context.Background(), keys)
	require.NoError(t, err)
	require.True(t, found[service.UsageLogDetailKey{RequestID: "req-1", APIKeyID: 2}])
	require.False(t, found[service.UsageLogDetailKey{RequestID: "req-2", APIKeyID: 3}])
	require.NoError(t, mock.ExpectationsWereMet())
}
