package handler

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/alexsey-popov/gkeep/internal/auth"
	"github.com/alexsey-popov/gkeep/internal/crypto"
	"github.com/alexsey-popov/gkeep/internal/models"
	"github.com/alexsey-popov/gkeep/internal/repository"
	"github.com/alexsey-popov/gkeep/internal/validator"
	"github.com/go-chi/jwtauth/v5"
	"github.com/stretchr/testify/require"
)

// mockRepository мок для репозитория
type mockRepository struct {
	users     map[string]*models.User
	passwords map[string]*models.Password
	cards     map[string]*models.Card
	texts     map[string]*models.Text
	files     map[string]*models.File

	// Ошибки для тестирования негативных сценариев работы БД
	userCreateErr     error
	userGetErr        error
	cardListErr       error
	cardCreateErr     error
	cardUpdateErr     error
	cardDeleteErr     error
	passwordListErr   error
	passwordCreateErr error
	passwordUpdateErr error
	passwordDeleteErr error
	textListErr       error
	textCreateErr     error
	textUpdateErr     error
	textDeleteErr     error
	fileListErr       error
	fileCreateErr     error
	fileShowErr       error
	fileDeleteErr     error
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		users:     make(map[string]*models.User),
		passwords: make(map[string]*models.Password),
		cards:     make(map[string]*models.Card),
		texts:     make(map[string]*models.Text),
		files:     make(map[string]*models.File),
	}
}

func (m *mockRepository) CreateUser(ctx context.Context, user *models.User) error {
	if m.userCreateErr != nil {
		return m.userCreateErr
	}
	if _, exists := m.users[user.Login]; exists {
		return repository.ErrConflictUnique
	}
	if user.Id == "" {
		user.Id = "00000000-0000-0000-0000-000000000001"
	}
	m.users[user.Login] = user
	return nil
}

func (m *mockRepository) GetUser(ctx context.Context, login string) (*models.User, error) {
	if m.userGetErr != nil {
		return nil, m.userGetErr
	}
	user, exists := m.users[login]
	if !exists {
		return nil, sql.ErrNoRows
	}
	return user, nil
}

func (m *mockRepository) CardList(ctx context.Context, userID, name, page string) ([]*models.Card, error) {
	if m.cardListErr != nil {
		return nil, m.cardListErr
	}
	var res []*models.Card
	for _, c := range m.cards {
		if c.UserId == userID {
			if name != "" && !strings.Contains(strings.ToLower(c.Name), strings.ToLower(name)) {
				continue
			}
			res = append(res, c)
		}
	}
	return res, nil
}

func (m *mockRepository) CardCreate(ctx context.Context, item *models.Card) error {
	if m.cardCreateErr != nil {
		return m.cardCreateErr
	}
	for _, c := range m.cards {
		if c.UserId == item.UserId && c.Name == item.Name {
			return repository.ErrConflictUnique
		}
	}
	if item.Id == "" {
		item.Id = fmt.Sprintf("card-%d", len(m.cards)+1)
	}
	item.CreatedAt = time.Now()
	m.cards[item.Id] = item
	return nil
}

func (m *mockRepository) CardUpdate(ctx context.Context, item *models.Card) error {
	if m.cardUpdateErr != nil {
		return m.cardUpdateErr
	}
	c, exists := m.cards[item.Id]
	if !exists || c.UserId != item.UserId {
		return sql.ErrNoRows
	}
	for _, other := range m.cards {
		if other.Id != item.Id && other.UserId == item.UserId && other.Name == item.Name {
			return repository.ErrConflictUnique
		}
	}
	c.Name = item.Name
	c.Number = item.Number
	c.Expired = item.Expired
	c.Code = item.Code
	return nil
}

func (m *mockRepository) CardDelete(ctx context.Context, userID, id string) (*models.Card, error) {
	if m.cardDeleteErr != nil {
		return nil, m.cardDeleteErr
	}
	c, exists := m.cards[id]
	if !exists || c.UserId != userID {
		return nil, sql.ErrNoRows
	}
	delete(m.cards, id)
	return c, nil
}

func (m *mockRepository) PasswordList(ctx context.Context, userID, name, page string) ([]*models.Password, error) {
	if m.passwordListErr != nil {
		return nil, m.passwordListErr
	}
	var res []*models.Password
	for _, p := range m.passwords {
		if p.UserId == userID {
			if name != "" && !strings.Contains(strings.ToLower(p.Name), strings.ToLower(name)) {
				continue
			}
			res = append(res, p)
		}
	}
	return res, nil
}

func (m *mockRepository) PasswordCreate(ctx context.Context, item *models.Password) error {
	if m.passwordCreateErr != nil {
		return m.passwordCreateErr
	}
	for _, p := range m.passwords {
		if p.UserId == item.UserId && p.Name == item.Name {
			return repository.ErrConflictUnique
		}
	}
	if item.Id == "" {
		item.Id = fmt.Sprintf("pwd-%d", len(m.passwords)+1)
	}
	item.CreatedAt = time.Now()
	m.passwords[item.Id] = item
	return nil
}

func (m *mockRepository) PasswordUpdate(ctx context.Context, item *models.Password) error {
	if m.passwordUpdateErr != nil {
		return m.passwordUpdateErr
	}
	p, exists := m.passwords[item.Id]
	if !exists || p.UserId != item.UserId {
		return sql.ErrNoRows
	}
	for _, other := range m.passwords {
		if other.Id != item.Id && other.UserId == item.UserId && other.Name == item.Name {
			return repository.ErrConflictUnique
		}
	}
	p.Name = item.Name
	p.Login = item.Login
	p.Password = item.Password
	return nil
}

func (m *mockRepository) PasswordDelete(ctx context.Context, userID, id string) (*models.Password, error) {
	if m.passwordDeleteErr != nil {
		return nil, m.passwordDeleteErr
	}
	p, exists := m.passwords[id]
	if !exists || p.UserId != userID {
		return nil, sql.ErrNoRows
	}
	delete(m.passwords, id)
	return p, nil
}

func (m *mockRepository) TextList(ctx context.Context, userID, name, page string) ([]*models.Text, error) {
	if m.textListErr != nil {
		return nil, m.textListErr
	}
	var res []*models.Text
	for _, t := range m.texts {
		if t.UserId == userID {
			if name != "" && !strings.Contains(strings.ToLower(t.Name), strings.ToLower(name)) {
				continue
			}
			res = append(res, t)
		}
	}
	return res, nil
}

func (m *mockRepository) TextCreate(ctx context.Context, item *models.Text) error {
	if m.textCreateErr != nil {
		return m.textCreateErr
	}
	for _, t := range m.texts {
		if t.UserId == item.UserId && t.Name == item.Name {
			return repository.ErrConflictUnique
		}
	}
	if item.Id == "" {
		item.Id = fmt.Sprintf("text-%d", len(m.texts)+1)
	}
	item.CreatedAt = time.Now()
	m.texts[item.Id] = item
	return nil
}

func (m *mockRepository) TextUpdate(ctx context.Context, item *models.Text) error {
	if m.textUpdateErr != nil {
		return m.textUpdateErr
	}
	t, exists := m.texts[item.Id]
	if !exists || t.UserId != item.UserId {
		return sql.ErrNoRows
	}
	for _, other := range m.texts {
		if other.Id != item.Id && other.UserId == item.UserId && other.Name == item.Name {
			return repository.ErrConflictUnique
		}
	}
	t.Name = item.Name
	t.Text = item.Text
	return nil
}

func (m *mockRepository) TextDelete(ctx context.Context, userID, id string) (*models.Text, error) {
	if m.textDeleteErr != nil {
		return nil, m.textDeleteErr
	}
	t, exists := m.texts[id]
	if !exists || t.UserId != userID {
		return nil, sql.ErrNoRows
	}
	delete(m.texts, id)
	return t, nil
}

func (m *mockRepository) FileList(ctx context.Context, userID, name, page string) ([]*models.File, error) {
	if m.fileListErr != nil {
		return nil, m.fileListErr
	}
	var res []*models.File
	for _, f := range m.files {
		if f.UserId == userID {
			if name != "" && !strings.Contains(strings.ToLower(f.Name), strings.ToLower(name)) {
				continue
			}
			res = append(res, f)
		}
	}
	return res, nil
}

func (m *mockRepository) FileCreate(ctx context.Context, item *models.File) error {
	if m.fileCreateErr != nil {
		return m.fileCreateErr
	}
	for _, f := range m.files {
		if f.UserId == item.UserId && f.Name == item.Name {
			return repository.ErrConflictUnique
		}
	}
	if item.Id == "" {
		item.Id = fmt.Sprintf("file-%d", len(m.files)+1)
	}
	item.CreatedAt = time.Now()
	m.files[item.Id] = item
	return nil
}

func (m *mockRepository) FileShow(ctx context.Context, userID, id string) (*models.File, error) {
	if m.fileShowErr != nil {
		return nil, m.fileShowErr
	}
	f, exists := m.files[id]
	if !exists || f.UserId != userID {
		return nil, sql.ErrNoRows
	}
	return f, nil
}

func (m *mockRepository) FileDelete(ctx context.Context, userID, id string) (*models.File, error) {
	if m.fileDeleteErr != nil {
		return nil, m.fileDeleteErr
	}
	f, exists := m.files[id]
	if !exists || f.UserId != userID {
		return nil, sql.ErrNoRows
	}
	delete(m.files, id)
	return f, nil
}

func setupTest(t *testing.T) (*Handler, *mockRepository, *auth.Guard, context.Context, string) {
	secret := "secret-jwt-token"
	guard := auth.New(secret)
	val, err := validator.New()
	require.NoError(t, err)

	cipher, err := crypto.NewWithTempKey()
	require.NoError(t, err)

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	rep := newMockRepository()
	h := New(rep, guard, cipher, val, log)

	userID := "00000000-0000-0000-0000-000000000001"
	tokenString, _, err := guard.NewUserToken(userID)
	require.NoError(t, err)

	token, err := guard.JWTAuth().Decode(tokenString)
	require.NoError(t, err)

	authCtx := jwtauth.NewContext(context.Background(), token, nil)

	return h, rep, guard, authCtx, userID
}
