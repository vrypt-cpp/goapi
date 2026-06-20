package users

import (
	"net/http"
	"sync"
	"sync/atomic"

	"goapi/core"
)

type Plugin struct {
	mu     sync.RWMutex
	store  map[int64]User
	nextID int64
}

func New() *Plugin {
	return &Plugin{store: map[int64]User{}}
}

func (p *Plugin) Name() string {
	return "users"
}

type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name" doc:"Display name of the user"`
	Email string `json:"email" doc:"Unique email address"`
	Role  string `json:"role" enum:"admin,member" doc:"User role"`
}

type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role" enum:"admin,member"`
}

type ListUsersResponse struct {
	Users []User `json:"users"`
}

func (p *Plugin) Routes() []core.Route {
	return []core.Route{
		{
			Method:   "GET",
			Path:     "/users",
			Summary:  "List users",
			Tags:     []string{"users"},
			Response: ListUsersResponse{},
			Handler:  p.list,
		},
		{
			Method:  "POST",
			Path:    "/users",
			Summary: "Create user",
			Tags:    []string{"users"},
			Request: CreateUserRequest{},
			Responses: map[int]any{
				201: User{},
				400: core.ErrorResponse{},
			},
			Handler: p.create,
		},
		{
			Method:  "GET",
			Path:    "/users/{id}",
			Summary: "Get user by ID",
			Tags:    []string{"users"},
			Params: []core.Parameter{
				{Name: "id", In: "path", Required: true, Schema: core.Schema{Type: "integer"}},
			},
			Responses: map[int]any{
				200: User{},
				404: core.ErrorResponse{},
			},
			Handler: p.get,
		},
		{
			Method:  "DELETE",
			Path:    "/users/{id}",
			Summary: "Delete user",
			Tags:    []string{"users"},
			Params: []core.Parameter{
				{Name: "id", In: "path", Required: true, Schema: core.Schema{Type: "integer"}},
			},
			Responses: map[int]any{
				204: nil,
				404: core.ErrorResponse{},
			},
			Handler: p.delete,
		},
	}
}

func (p *Plugin) list(w http.ResponseWriter, r *http.Request) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	out := make([]User, 0, len(p.store))
	for _, u := range p.store {
		out = append(out, u)
	}
	core.JSON(w, http.StatusOK, ListUsersResponse{Users: out})
}

func (p *Plugin) create(w http.ResponseWriter, r *http.Request) {
	req, err := core.Decode[CreateUserRequest](r)
	if err != nil {
		core.JSON(w, http.StatusBadRequest, core.ErrorResponse{Error: err.Error()})
		return
	}
	id := atomic.AddInt64(&p.nextID, 1)
	user := User{ID: id, Name: req.Name, Email: req.Email, Role: req.Role}

	p.mu.Lock()
	p.store[id] = user
	p.mu.Unlock()

	core.JSON(w, http.StatusCreated, user)
}

func (p *Plugin) get(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.PathValue("id"))
	p.mu.RLock()
	user, ok := p.store[id]
	p.mu.RUnlock()
	if !ok {
		core.JSON(w, http.StatusNotFound, core.ErrorResponse{Error: "user not found"})
		return
	}
	core.JSON(w, http.StatusOK, user)
}

func (p *Plugin) delete(w http.ResponseWriter, r *http.Request) {
	id := parseID(r.PathValue("id"))
	p.mu.Lock()
	_, ok := p.store[id]
	delete(p.store, id)
	p.mu.Unlock()
	if !ok {
		core.JSON(w, http.StatusNotFound, core.ErrorResponse{Error: "user not found"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func parseID(s string) int64 {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int64(c-'0')
	}
	return n
}
