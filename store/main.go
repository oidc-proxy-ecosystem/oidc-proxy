package store

import (
	"context"
	"encoding/base32"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/securecookie"
	"github.com/gorilla/sessions"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/session"
)

var sessionExpire = 86400 * 30

type SessionStore struct {
	session    session.Session
	Options    *sessions.Options
	StoreMutex sync.RWMutex
	keyPairs   []securecookie.Codec
}

func (store *SessionStore) Close() error {
	if store.session != nil {
		return store.session.Close(context.Background())
	}
	return nil
}

type sessionValues map[interface{}]interface{}

func (s sessionValues) mapToJson() (string, error) {
	mp := map[string]interface{}{}
	for key, value := range s {
		mp[key.(string)] = value
	}
	buf, err := json.Marshal(&mp)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func (s *sessionValues) jsonToMap(str string) error {
	values := sessionValues{}
	mp := map[string]interface{}{}
	if err := json.Unmarshal([]byte(str), &mp); err != nil {
		return err
	}
	for key, val := range mp {
		values[key] = val
	}
	*s = values
	return nil
}

var _ sessions.Store = &SessionStore{}

func NewStore(c session.Session, codec ...[]byte) *SessionStore {
	store := &SessionStore{
		session: c,
		Options: &sessions.Options{
			MaxAge: sessionExpire,
			Secure: true,
		},
		keyPairs: securecookie.CodecsFromPairs(codec...),
	}
	return store
}

func getCancelContext() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (store *SessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return sessions.GetRegistry(r).Get(store, name)
}

func (store *SessionStore) New(r *http.Request, name string) (*sessions.Session, error) {
	var err error
	session := sessions.NewSession(store, name)
	opts := *store.Options
	session.Options = &opts
	session.IsNew = true
	if cookie, errCookie := r.Cookie(name); errCookie == nil {
		err = securecookie.DecodeMulti(name, cookie.Value, &session.ID, store.keyPairs...)
		if err == nil {
			err := store.load(session)
			session.IsNew = !(err == nil)
		}
	}
	return session, err
}

func (store *SessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	if session.Options.MaxAge < 0 {
		if err := store.Delete(session); err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), "", session.Options))
	} else {
		if session.ID == "" {
			session.ID = strings.TrimRight(base32.StdEncoding.EncodeToString(securecookie.GenerateRandomKey(32)), "=")
		}
		if err := store.save(session); err != nil {
			return err
		}
		encoded, err := securecookie.EncodeMulti(session.Name(), session.ID, store.keyPairs...)
		if err != nil {
			return err
		}
		http.SetCookie(w, sessions.NewCookie(session.Name(), encoded, session.Options))
		// http.SetCookie(w, &http.Cookie{
		// 	Name:  session.Name(),
		// 	Value: encoded,
		// })
	}
	return nil
}

func (store *SessionStore) save(session *sessions.Session) error {
	value := sessionValues(session.Values)
	encoded, err := value.mapToJson()
	if err != nil {
		return err
	}
	ctx, cancel := getCancelContext()
	defer cancel()
	store.StoreMutex.Lock()
	defer store.StoreMutex.Unlock()
	key := "session_" + session.ID
	return store.session.Put(ctx, key, encoded)
}

func (store *SessionStore) load(session *sessions.Session) error {
	values := sessionValues{}
	ctx, cancel := getCancelContext()
	defer cancel()
	store.StoreMutex.Lock()
	defer store.StoreMutex.Unlock()
	key := "session_" + session.ID
	value, err := store.session.Get(ctx, key)
	if err != nil {
		return err
	}
	err = values.jsonToMap(value)
	if err != nil {
		return err
	}
	session.Values = values
	return nil
}

func (store *SessionStore) Delete(session *sessions.Session) error {
	ctx, cancel := getCancelContext()
	defer cancel()
	store.StoreMutex.Lock()
	defer store.StoreMutex.Unlock()
	key := "session_" + session.ID
	return store.session.Delete(ctx, key)
}
