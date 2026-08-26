package session

import "context"

type Manager struct {
	store *Store
}

func NewManager(store *Store) *Manager {
	return &Manager{store: store}
}

func (m *Manager) Open(ctx context.Context, sess Session) error {
	return m.store.SaveSession(ctx, &sess)
}

func (m *Manager) Close(ctx context.Context, id string, status SessionStatus) error {
	sess, err := m.store.LoadSession(ctx, id)
	if err != nil {
		return err
	}
	sess.Status = status
	return m.store.SaveSession(ctx, sess)
}

func (m *Manager) Get(ctx context.Context, id string) (*Session, error) {
	return m.store.LoadSession(ctx, id)
}

func (m *Manager) List(ctx context.Context) ([]Session, error) {
	return m.store.ListSessions(ctx)
}

func (m *Manager) AppendEvent(ctx context.Context, sessionID string, e StoredEvent) error {
	e.SessionID = sessionID
	return m.store.AppendEvent(ctx, &e)
}

func (m *Manager) LoadEvents(ctx context.Context, sessionID string) ([]StoredEvent, error) {
	return m.store.LoadEvents(ctx, sessionID)
}
