package mock

import (
	"context"
	"sync"

	"github.com/erknas/ecom/user-service/internal/domain/models"
	"github.com/erknas/ecom/user-service/internal/storage"
)

type MockStorage struct {
	mu              sync.RWMutex
	users           map[int64]*models.User
	emailUniqueness map[string]int64
	nextID          int64
}

func NewMockStorage() *MockStorage {
	return &MockStorage{
		users:           make(map[int64]*models.User),
		emailUniqueness: make(map[string]int64),
		nextID:          1,
	}
}

func (m *MockStorage) Insert(ctx context.Context, user *models.User) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return 0, ctx.Err()
	default:
	}

	if _, exists := m.emailUniqueness[user.Email]; exists {
		return 0, storage.ErrUserExists
	}

	id := m.nextID
	m.nextID++

	userCopy := &models.User{
		ID:           id,
		FirstName:    user.FirstName,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}

	m.users[id] = userCopy
	m.emailUniqueness[user.Email] = id

	return id, nil
}

func (m *MockStorage) UserByID(ctx context.Context, id int64) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	user, exists := m.users[id]
	if !exists {
		return nil, storage.ErrUserNotFound
	}

	return &models.User{
		ID:           id,
		FirstName:    user.FirstName,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (m *MockStorage) UserByEmail(ctx context.Context, email string) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	id, exists := m.emailUniqueness[email]
	if !exists {
		return nil, storage.ErrUserNotFound
	}

	user := m.users[id]

	return &models.User{
		ID:           id,
		FirstName:    user.FirstName,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    user.CreatedAt,
	}, nil
}

func (m *MockStorage) Update(ctx context.Context, id int64, user *models.UpdatedUser) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	existing, exists := m.users[id]
	if !exists {
		return storage.ErrUserNotFound
	}

	if user.FirstName == nil && user.Email == nil && user.PasswordHash == nil {
		return storage.ErrNoChanges
	}

	if user.Email != nil && *user.Email != existing.Email {
		if _, emailExists := m.emailUniqueness[*user.Email]; emailExists {
			return storage.ErrUserExists
		}
	}

	updatedUser := &models.User{
		ID:           existing.ID,
		FirstName:    existing.FirstName,
		Email:        existing.Email,
		PasswordHash: existing.PasswordHash,
		CreatedAt:    existing.CreatedAt,
	}

	if user.FirstName != nil {
		updatedUser.FirstName = *user.FirstName
	}

	if user.Email != nil {
		delete(m.emailUniqueness, existing.Email)
		m.emailUniqueness[*user.Email] = id
		updatedUser.Email = *user.Email
	}

	if user.PasswordHash != nil {
		updatedUser.PasswordHash = user.PasswordHash
	}

	m.users[id] = updatedUser

	return nil
}
