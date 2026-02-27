package user

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type mockRepo struct {
	getByEmail func(ctx context.Context, email string) (User, error)
	create     func(ctx context.Context, username, email, passwordHash, firstName, lastName string) (User, error)
}

func (m *mockRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	if m.getByEmail != nil {
		return m.getByEmail(ctx, email)
	}
	return User{}, errors.New("not found")
}

func (m *mockRepo) Create(ctx context.Context, username, email, passwordHash, firstName, lastName string) (User, error) {
	if m.create != nil {
		return m.create(ctx, username, email, passwordHash, firstName, lastName)
	}
	return User{}, nil
}

func (m *mockRepo) GetByID(ctx context.Context, id string) (User, error) {
	return User{}, nil
}

func (m *mockRepo) List(ctx context.Context, limit, offset int) ([]User, error) {
	return nil, nil
}

func TestRegister(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		var savedHash string
		r := &mockRepo{
			getByEmail: func(context.Context, string) (User, error) { return User{}, errors.New("") },
			create: func(_ context.Context, _, _, hash, _, _ string) (User, error) {
				savedHash = hash
				return User{ID: "1", Email: "a@b"}, nil
			},
		}
		u, err := NewService(r).Register(context.Background(), "a@b", "qwerty")
		if err != nil {
			t.Fatal(err)
		}
		if u.Email != "a@b" {
			t.Fail()
		}
		if savedHash == "qwerty" {
			t.Fatal("password must be hashed")
		}
	})

	t.Run("duplicate email", func(t *testing.T) {
		r := &mockRepo{
			getByEmail: func(_ context.Context, email string) (User, error) { return User{Email: email}, nil },
		}
		_, err := NewService(r).Register(context.Background(), "same@x.com", "x")
		if err != ErrEmailExists {
			t.Fatalf("got %v, want ErrEmailExists", err)
		}
	})
}

func TestLogin(t *testing.T) {
	okHash, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.DefaultCost)

	t.Run("success", func(t *testing.T) {
		r := &mockRepo{
			getByEmail: func(_ context.Context, email string) (User, error) {
				return User{ID: "1", Email: email, PasswordHash: string(okHash)}, nil
			},
		}
		u, err := NewService(r).Login(context.Background(), "u@x.com", "right")
		if err != nil {
			t.Fatal(err)
		}
		if u.Email != "u@x.com" {
			t.Fail()
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		r := &mockRepo{
			getByEmail: func(_ context.Context, email string) (User, error) {
				return User{PasswordHash: string(okHash)}, nil
			},
		}
		_, err := NewService(r).Login(context.Background(), "u@x.com", "wrong")
		if err != ErrInvalidCreds {
			t.Fatalf("got %v, want ErrInvalidCreds", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		r := &mockRepo{getByEmail: func(context.Context, string) (User, error) { return User{}, errors.New("") }}
		_, err := NewService(r).Login(context.Background(), "ghost@x.com", "any")
		if err != ErrInvalidCreds {
			t.Fatalf("got %v, want ErrInvalidCreds", err)
		}
	})
}
