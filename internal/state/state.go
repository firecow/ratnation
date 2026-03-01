// Package state provides types and utilities for managing council state.
package state

// State represents the full council state including services, kings, and lings.
type State struct {
	Revision int       `json:"revision"`
	Services []Service `json:"services"`
	Kings    []King    `json:"kings"`
	Lings    []Ling    `json:"lings"`
}

// King represents a king node in the council state.
type King struct {
	BindPort     int    `json:"bindPort"`
	Host         string `json:"host"`
	Ports        string `json:"ports"`
	ShuttingDown bool   `json:"shuttingDown"`
	Beat         int64  `json:"beat"`
	Location     string `json:"location"`
	CertPEM      string `json:"certPem,omitempty"`
}

// Ling represents a ling node in the council state.
type Ling struct {
	LingID       string `json:"lingId"`
	ShuttingDown bool   `json:"shuttingDown"`
	Beat         int64  `json:"beat"`
}

// Service represents a service entry in the council state.
type Service struct {
	Name              string  `json:"name"`
	Token             string  `json:"token"`
	ServiceID         string  `json:"serviceId"`
	LingID            string  `json:"lingId"`
	PreferredLocation string  `json:"preferredLocation"`
	LingReady         bool    `json:"lingReady"`
	KingReady         bool    `json:"kingReady"`
	Host              *string `json:"host"`
	BindPort          *int    `json:"bindPort"`
	RemotePort        *int    `json:"remotePort"`
}
