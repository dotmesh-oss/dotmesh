package user

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dotmesh-oss/dotmesh/pkg/types"

	log "github.com/sirupsen/logrus"
)

type ExternalManager struct {
	url        string
	httpClient *http.Client
}

func NewExternal(url string, httpClient *http.Client) *ExternalManager {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	return &ExternalManager{
		url:        url,
		httpClient: httpClient,
	}
}

func (m *ExternalManager) call(operation string, method string, body interface{}, result interface{}) error {
	l := log.WithFields(log.Fields{
		"base_url":  m.url,
		"operation": operation,
		"method":    method,
		"body":      fmt.Sprintf("%#v", body),
	})

	var err error
	for retries := 60; retries > 0; retries-- {
		var bodyReader io.Reader
		if body != nil {
			bodyEncoded, err := json.Marshal(body)
			if err != nil {
				l.WithError(err).Error("[externalManager] Error encoding body")
				// No sense retrying this
				return err
			}
			bodyReader = bytes.NewBuffer(bodyEncoded)
		} else {
			bodyReader = nil
		}

		req, err := http.NewRequest(method, m.url+"/"+operation, bodyReader)
		if err != nil {
			l.WithError(err).Error("[externalManager] Error creating HTTP request")
			// No sense retrying this
			return err
		}

		req.Header.Set("Content-Type", "application/json")

		resp, err := m.httpClient.Do(req)
		if err != nil {
			l.WithError(err).Error("[externalManager] Error performing HTTP request")
			// Retry
			time.Sleep(time.Second)
			continue
		}
		defer resp.Body.Close()

		switch resp.StatusCode {
		case 200, 201:
			// All is well, proceed
		default:
			l.WithField("http_status", resp.Status).Error("[externalManager] HTTP error")

			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				l.WithError(err).Error("[externalManager] Error reading response body")
				// Retry
				time.Sleep(time.Second)
				continue
			}

			// No sense retrying if we got a reply, but it was an error
			return fmt.Errorf("HTTP Error: %d body: %q", resp.StatusCode, string(b))
		}

		if result != nil {
			b, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				l.WithError(err).Error("[externalManager] Error reading response body")
				// Retry
				time.Sleep(time.Second)
				continue
			}

			err = json.Unmarshal(b, result)
			if err != nil {
				l.WithError(err).Error("[externalManager] Error decoding response body")
				// Retry
				time.Sleep(time.Second)
				continue
			}

			//		l.WithField("response", fmt.Sprintf("%#v", result)).Debug("ABS DEBUG response")
		}

		return nil
	}

	// We keep the last error encountered, to return it if the retry loop runs out
	return err
}

func (m *ExternalManager) NewAdmin(user *User) error {
	return m.call("user/admin", http.MethodPut, user, nil)
}

type NewUserRequest struct {
	Name     string
	Email    string
	Password string
}

func (m *ExternalManager) New(name, email, password string) (*User, error) {
	var u User
	err := m.call("user", http.MethodPut, NewUserRequest{
		Name:     name,
		Email:    email,
		Password: password,
	}, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (m *ExternalManager) Get(q *Query) (*User, error) {
	var u User
	err := m.call("user", http.MethodGet, q, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (m *ExternalManager) Update(user *User) (*User, error) {
	var u User
	err := m.call("user", http.MethodPost, user, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

// Import user without hashing password or generating API key
func (m *ExternalManager) Import(user *User) error {
	return m.call("user/import", http.MethodPut, user, nil)
}

type UpdatePasswordRequest struct {
	UserID      string
	NewPassword string
}

func (m *ExternalManager) UpdatePassword(id string, password string) (*User, error) {
	var u User
	err := m.call("user/password", http.MethodPost, UpdatePasswordRequest{
		UserID:      id,
		NewPassword: password,
	}, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

type ResetAPIKeyRequest struct {
	UserID string
}

func (m *ExternalManager) ResetAPIKey(id string) (*User, error) {
	var u User
	err := m.call("user/api-key", http.MethodPost, ResetAPIKeyRequest{
		UserID: id,
	}, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

type DeleteRequest struct {
	UserID string
}

func (m *ExternalManager) Delete(id string) error {
	return m.call("user", http.MethodDelete, DeleteRequest{
		UserID: id,
	}, nil)
}

type ListRequest struct {
	Selector string
}

func (m *ExternalManager) List(selector string) ([]*User, error) {
	var u []*User
	err := m.call("user/list", http.MethodGet, ListRequest{
		Selector: selector,
	}, &u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

type AuthenticateRequest struct {
	Username string
	Password string
}

type AuthenticateResponse struct {
	User User
	Type string
}

func (m *ExternalManager) Authenticate(username, password string) (*User, AuthenticationType, error) {
	var ar AuthenticateResponse
	err := m.call("user/authenticate", http.MethodPost, AuthenticateRequest{
		Username: username,
		Password: password,
	}, &ar)
	if err != nil {
		return nil, AuthenticationTypeNone, err
	}

	at, err := AuthenticationTypeFromString(ar.Type)
	if err != nil {
		return nil, AuthenticationTypeNone, err
	}

	return &(ar.User), at, nil
}

type AuthorizeRequest struct {
	User               User
	OwnerAction        bool
	TopLevelFilesystem types.TopLevelFilesystem
}

type AuthorizeResponse struct {
	Allowed bool
}

func (m *ExternalManager) Authorize(user *User, ownerAction bool, tlf *types.TopLevelFilesystem) (bool, error) {
	var ar AuthorizeResponse
	err := m.call("authorize", http.MethodPost, AuthorizeRequest{
		User:               *user,
		OwnerAction:        ownerAction,
		TopLevelFilesystem: *tlf,
	}, &ar)
	if err != nil {
		return false, err
	}
	return ar.Allowed, nil
}

type AuthorizeNamespaceAdminRequest struct {
	User      User
	Namespace string
}

func (m *ExternalManager) UserIsNamespaceAdministrator(user *User, namespace string) (bool, error) {
	var ar AuthorizeResponse
	err := m.call("authorize-namespace-admin", http.MethodPost, AuthorizeNamespaceAdminRequest{
		User:      *user,
		Namespace: namespace,
	}, &ar)
	if err != nil {
		return false, err
	}
	return ar.Allowed, nil
}
