package model

import "k8s.io/apimachinery/pkg/util/json"

type EnvelopeWithPayload[T any] struct {
	Type           string `json:"type,omitempty"`
	MessageId      string `json:"messageId,omitempty"`
	EventName      string `json:"eventName"`
	Version        string `json:"version"`
	XCorrelationId string `json:"x-correlationId"`
	XSender        string `json:"x-sender"`
	Payload        T      `json:"payload"`
}

func SerialiseToEnvelopeWithPayload[T any](data []byte) (*EnvelopeWithPayload[T], error) {
	var envelope EnvelopeWithPayload[T]
	err := json.Unmarshal(data, &envelope)
	return &envelope, err
}
