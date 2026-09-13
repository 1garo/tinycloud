package app

import "time"

type Status string

const (
	StatusCreated  Status = "created"
	StatusBuilding Status = "building"
	StatusRunning  Status = "running"
	StatusStopped  Status = "stopped"
	StatusFailed   Status = "failed"
)

type App struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	SourceDir     string    `json:"source_dir"`
	ContainerPort string    `json:"container_port"`
	HostPort      string    `json:"host_port"`
	ContainerID   string    `json:"container_id,omitempty"`
	Status        Status    `json:"status"`
	LastError     string    `json:"last_error,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (a App) URL() string {
	return "http://" + a.Name + ".localhost:" + a.HostPort
}
