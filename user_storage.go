package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/boltdb/bolt"
)

// UserStorage defines the interface for user storage
type UserStorage interface {
	Add(user *User) error
	Get(id int) (*User, error)
	GetAll() ([]*User, error)
	Update(user *User) error
	Delete(id int) error
	GetByUsername(username string) (*User, error)
	GetByEmail(email string) (*User, error)
}

// BoltDBUserStorage implements UserStorage using BoltDB
type BoltDBUserStorage struct {
	db         *bolt.DB
	bucketName []byte
	mu         sync.RWMutex
	nextID     int
}

// NewBoltDBUserStorage creates a new BoltDB user storage
func NewBoltDBUserStorage(db *bolt.DB) (*BoltDBUserStorage, error) {
	bucketName := []byte("users")
	
	// Initialize bucket
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		return err
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to create users bucket: %v", err)
	}
	
	// Find the highest ID
	nextID := 1
	err = db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)
		c := b.Cursor()
		
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			id := btoi(k)
			if id >= nextID {
				nextID = id + 1
			}
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to determine next user ID: %v", err)
	}
	
	return &BoltDBUserStorage{
		db:         db,
		bucketName: bucketName,
		nextID:     nextID,
	}, nil
}

// Add saves a new user
func (s *BoltDBUserStorage) Add(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		
		// Assign ID
		user.ID = s.nextID
		s.nextID++
		
		// Serialize user
		buf, err := json.Marshal(user)
		if err != nil {
			return err
		}
		
		// Save to bucket
		return b.Put(itob(user.ID), buf)
	})
}

// Get retrieves a user by ID
func (s *BoltDBUserStorage) Get(id int) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var user *User
	
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		data := b.Get(itob(id))
		
		if data == nil {
			return nil // User not found
		}
		
		// Deserialize user
		user = &User{}
		return json.Unmarshal(data, user)
	})
	
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetAll returns all users
func (s *BoltDBUserStorage) GetAll() ([]*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	users := make([]*User, 0)
	
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		
		return b.ForEach(func(k, v []byte) error {
			user := &User{}
			if err := json.Unmarshal(v, user); err != nil {
				return err
			}
			
			users = append(users, user)
			return nil
		})
	})
	
	if err != nil {
		return nil, err
	}
	
	return users, nil
}

// Update modifies an existing user
func (s *BoltDBUserStorage) Update(user *User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		
		// Check if user exists
		if b.Get(itob(user.ID)) == nil {
			return errors.New("user not found")
		}
		
		// Serialize user
		buf, err := json.Marshal(user)
		if err != nil {
			return err
		}
		
		// Update in bucket
		return b.Put(itob(user.ID), buf)
	})
}

// Delete removes a user by ID
func (s *BoltDBUserStorage) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		return b.Delete(itob(id))
	})
}

// GetByUsername finds a user by username
func (s *BoltDBUserStorage) GetByUsername(username string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var foundUser *User
	
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		
		return b.ForEach(func(k, v []byte) error {
			user := &User{}
			if err := json.Unmarshal(v, user); err != nil {
				return err
			}
			
			if user.Username == username {
				foundUser = user
				return errors.New("user found") // Use error to break the loop
			}
			
			return nil
		})
	})
	
	if err != nil && err.Error() != "user found" {
		return nil, err
	}
	
	return foundUser, nil
}

// GetByEmail finds a user by email
func (s *BoltDBUserStorage) GetByEmail(email string) (*User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var foundUser *User
	
	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(s.bucketName)
		
		return b.ForEach(func(k, v []byte) error {
			user := &User{}
			if err := json.Unmarshal(v, user); err != nil {
				return err
			}
			
			if user.Email == email {
				foundUser = user
				return errors.New("user found") // Use error to break the loop
			}
			
			return nil
		})
	})
	
	if err != nil && err.Error() != "user found" {
		return nil, err
	}
	
	return foundUser, nil
}
