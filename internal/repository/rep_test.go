package repository

import (
	"log/slog"
	"os"
	"testing"

	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDB(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	cipher, err := crypto.NewWithTempKey()
	require.NoError(t, err)

	db := NewDB(nil, logger, cipher)

	require.NotNil(t, db)
	assert.Equal(t, logger, db.log)
	assert.Nil(t, db.db)
	assert.Equal(t, cipher, db.cipher)
}
