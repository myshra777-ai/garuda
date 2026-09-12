// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0
//
// Law Enforcement. I am bound by the ACGM Resolution Invariant and the 10 Immutable Laws. Truth Preservation is Absolute.

package api

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

// SessionCookieName is the name of the session cookie set by the login
// handler. Read by the session middleware on every protected request.
const SessionCookieName = "garuda_session"

// SessionTTL is the inactivity timeout for a browser session.
// 24 hours matches Grafana's default for small deployments.
const SessionTTL = 24 * time.Hour

const sessionGCLoopInterval = 5 * time.Minute

// Session represents an authenticated browser session.
//
// The session identifies a user. It does not carry a tenant — that
// will be added when the users table gains a tenant_id column. For
// now every user is in the single default tenant.
type Session struct {
	ID        string
	UserID    string
	UserEmail string
	Role      string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionStore is an in-memory map of session IDs to sessions.
//
// Sessions do not persist across daemon restarts. If the daemon
// restarts, every logged-in browser has to log in again. For a
// single-node deployment this is correct and simple. When Garuda
// supports a Redis-backed multi-node deployment, this interface
// stays the same and the implementation changes.
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	ttl      time.Duration
	stopGC   chan struct{}
}

// NewSessionStore creates a store and starts the background GC loop.
func NewSessionStore(ttl time.Duration) *SessionStore {
	s := &SessionStore{
		sessions: make(map[string]*Session),
		ttl:      ttl,
		stopGC:   make(chan struct{}),
	}
	go s.gcLoop()
	return s
}

// Create generates a new session for the given user.
func (s *SessionStore) Create(userID, userEmail, role string) *Session {
	now := time.Now().UTC()
	sess := &Session{
		ID:        generateSessionID(),
		UserID:    userID,
		UserEmail: userEmail,
		Role:      role,
		CreatedAt: now,
		ExpiresAt: now.Add(s.ttl),
	}
	s.mu.Lock()
	s.sessions[sess.ID] = sess
	s.mu.Unlock()
	return sess
}

// Get retrieves a session by ID. Expired sessions are deleted on access.
func (s *SessionStore) Get(id string) (*Session, bool) {
	s.mu.RLock()
	sess, ok := s.sessions[id]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().UTC().After(sess.ExpiresAt) {
		s.Delete(id)
		return nil, false
	}
	return sess, true
}

// Delete removes a session. Called on logout and on GC.
func (s *SessionStore) Delete(id string) {
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

// Close stops the background GC loop. Call from daemon shutdown.
func (s *SessionStore) Close() {
	close(s.stopGC)
}

func (s *SessionStore) gcLoop() {
	ticker := time.NewTicker(sessionGCLoopInterval)
	defer ticker.Stop()
	for {
		select {
		case <-s.stopGC:
			return
		case <-ticker.C:
			now := time.Now().UTC()
			s.mu.Lock()
			for id, sess := range s.sessions {
				if now.After(sess.ExpiresAt) {
					delete(s.sessions, id)
				}
			}
			s.mu.Unlock()
		}
	}
}

func generateSessionID() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand failing is a fatal system condition; there is no
		// safe fallback. Panic so the daemon does not start with a
		// predictable session ID space.
		panic("session: crypto/rand unavailable: " + err.Error())
	}
	return hex.EncodeToString(b)
}
