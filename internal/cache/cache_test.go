package cache

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCache_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "test.cache")

	key := GetKey("user123", "login123")
	c, err := New(cachePath, key)
	require.NoError(t, err)

	// Добавим данные каждого типа
	passwords := []models.Password{
		{Id: "p1", Name: "Email", Login: "user@mail.com", Password: "secretPassword"},
	}
	cards := []models.Card{
		{Id: "c1", Name: "Visa", Number: "1111222233334444", Expired: "12/28", Code: "123"},
	}
	texts := []models.Text{
		{Id: "t1", Name: "Note", Text: "Hello world"},
	}
	files := []models.File{
		{Id: "f1", Name: "Doc", Filename: "doc.txt", Size: 1024, Checksum: "abc"},
	}

	c.SetPasswords(1, passwords)
	c.SetCards(1, cards)
	c.SetTexts(1, texts)
	c.SetFiles(1, files)

	err = c.Save()
	require.NoError(t, err)

	// Загружаем в новый объект Cache с тем же ключом
	c2, err := New(cachePath, key)
	require.NoError(t, err)

	err = c2.Load()
	require.NoError(t, err)

	loadedPasswords, ok := c2.GetPasswords(1)
	assert.True(t, ok)
	assert.Equal(t, passwords, loadedPasswords)

	loadedCards, ok := c2.GetCards(1)
	assert.True(t, ok)
	assert.Equal(t, cards, loadedCards)

	loadedTexts, ok := c2.GetTexts(1)
	assert.True(t, ok)
	assert.Equal(t, texts, loadedTexts)

	loadedFiles, ok := c2.GetFiles(1)
	assert.True(t, ok)
	assert.Equal(t, files, loadedFiles)

	_, ok = c2.GetPasswords(2)
	assert.False(t, ok)
}

func TestCache_LoadNonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "non_existent.cache")

	key := GetKey("user123", "login123")
	c, err := New(cachePath, key)
	require.NoError(t, err)

	err = c.Load()
	assert.NoError(t, err)
}

func TestCache_LoadWrongKey(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "wrong_key.cache")

	key1 := GetKey("user1", "login1")
	key2 := GetKey("user2", "login2")

	c1, err := New(cachePath, key1)
	require.NoError(t, err)

	c1.SetPasswords(1, []models.Password{{Id: "p1", Name: "Email"}})
	err = c1.Save()
	require.NoError(t, err)

	c2, err := New(cachePath, key2)
	require.NoError(t, err)

	err = c2.Load()
	assert.ErrorIs(t, err, ErrDecrypt)
}

func TestCache_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()
	cachePath := filepath.Join(tmpDir, "empty.cache")
	require.NoError(t, os.WriteFile(cachePath, []byte{}, 0600))

	key := GetKey("user1", "login1")
	c, err := New(cachePath, key)
	require.NoError(t, err)

	err = c.Load()
	assert.NoError(t, err)
}
