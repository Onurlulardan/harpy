package runtimev1

import (
	"encoding/json"
	"fmt"
)

type LivenessStatus string

const LivenessStatusAlive LivenessStatus = "alive"

type Livez struct {
	Status  LivenessStatus `json:"status"`
	Service Service        `json:"service"`
}

func NewLivez(service Service) (Livez, error) {
	livez := Livez{Status: LivenessStatusAlive, Service: service}
	if err := livez.Validate(); err != nil {
		return Livez{}, err
	}
	return livez, nil
}

func (livez Livez) Validate() error {
	if livez.Status != LivenessStatusAlive {
		return fmt.Errorf("liveness status must be alive")
	}
	if !livez.Service.Valid() {
		return fmt.Errorf("liveness service is invalid")
	}
	return nil
}

func (livez Livez) MarshalJSON() ([]byte, error) {
	if err := livez.Validate(); err != nil {
		return nil, err
	}
	type wireLivez Livez
	return json.Marshal(wireLivez(livez))
}
