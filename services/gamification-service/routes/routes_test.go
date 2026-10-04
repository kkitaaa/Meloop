package routes

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/meloop/gamification-service/models"
)

type rewardManagerFake struct {
	reward models.Reward
}

func (manager *rewardManagerFake) List(context.Context, bool) ([]models.Reward, error) {
	return []models.Reward{manager.reward}, nil
}

func (manager *rewardManagerFake) Get(context.Context, string) (*models.Reward, error) {
	return &manager.reward, nil
}

func (manager *rewardManagerFake) Create(_ context.Context, request models.RewardRequest) (*models.Reward, error) {
	manager.reward = models.Reward{ID: "reward-1", Name: request.Name, Description: request.Description, Type: request.Type, RequiredLevel: request.RequiredLevel, Available: true}
	if request.Available != nil {
		manager.reward.Available = *request.Available
	}
	return &manager.reward, nil
}

func (manager *rewardManagerFake) Update(_ context.Context, id string, request models.RewardRequest) (*models.Reward, error) {
	manager.reward = models.Reward{ID: id, Name: request.Name, Description: request.Description, Type: request.Type, RequiredLevel: request.RequiredLevel}
	if request.Available != nil {
		manager.reward.Available = *request.Available
	}
	return &manager.reward, nil
}

func (manager *rewardManagerFake) SetAvailable(_ context.Context, id string, available bool) (*models.Reward, error) {
	manager.reward.ID = id
	manager.reward.Available = available
	return &manager.reward, nil
}

type roleRepositoryFake bool

func (repository roleRepositoryFake) IsAdmin(context.Context, string) (bool, error) {
	return bool(repository), nil
}

func TestRewardCRUDRequiresAdministrator(t *testing.T) {
	for _, test := range []struct {
		name       string
		token      string
		isAdmin    bool
		wantStatus int
	}{
		{name: "no token", wantStatus: http.StatusUnauthorized},
		{name: "valid user is not admin", token: "Bearer user-token", wantStatus: http.StatusForbidden},
		{name: "administrator", token: "Bearer admin-token", isAdmin: true, wantStatus: http.StatusCreated},
	} {
		t.Run(test.name, func(t *testing.T) {
			authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
				if request.Header.Get("Authorization") != "Bearer admin-token" && request.Header.Get("Authorization") != "Bearer user-token" {
					http.Error(w, "unauthorized", http.StatusUnauthorized)
					return
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"success":true,"data":{"id":"session-user"}}`))
			}))
			defer authServer.Close()
			router := Setup(&rewardManagerFake{}, authServer.URL, roleRepositoryFake(test.isAdmin))
			request := httptest.NewRequest(http.MethodPost, "/admin/rewards", strings.NewReader(`{"nombre":"Badge","tipo":"BADGE","nivel_requerido":2}`))
			request.Header.Set("Content-Type", "application/json")
			if test.token != "" {
				request.Header.Set("Authorization", test.token)
			}
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d: %s", response.Code, test.wantStatus, response.Body.String())
			}
		})
	}
}

func TestAdministratorCanCreateAndDisableReward(t *testing.T) {
	authServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"data":{"id":"admin-1"}}`))
	}))
	defer authServer.Close()
	manager := &rewardManagerFake{}
	router := Setup(manager, authServer.URL, roleRepositoryFake(true))

	request := httptest.NewRequest(http.MethodPost, "/admin/rewards", strings.NewReader(`{"nombre":"Badge","descripcion":"Primer logro","tipo":"BADGE","nivel_requerido":2}`))
	request.Header.Set("Authorization", "Bearer admin-token")
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d: %s", response.Code, http.StatusCreated, response.Body.String())
	}

	request = httptest.NewRequest(http.MethodDelete, "/admin/rewards/reward-1", nil)
	request.Header.Set("Authorization", "Bearer admin-token")
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("delete status = %d, want %d: %s", response.Code, http.StatusOK, response.Body.String())
	}
	if manager.reward.Available {
		t.Fatal("DELETE should disable the reward without deleting it")
	}
	var envelope struct {
		Success bool `json:"success"`
		Data    struct {
			Available bool `json:"disponible"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !envelope.Success || envelope.Data.Available {
		t.Fatalf("unexpected disable response: %+v", envelope)
	}
}
