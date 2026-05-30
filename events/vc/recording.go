// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package vc

import (
	"context"
	"encoding/json"

	"github.com/larksuite/cli/internal/event"
)

// VCRecordingTranscriptItemOutput is one flattened transcript item for recording events.
type VCRecordingTranscriptItemOutput struct {
	Speaker     *VCRecordingSpeakerOutput `json:"speaker,omitempty"       desc:"Speaker metadata"`
	Text        string                    `json:"text,omitempty"          desc:"Transcript text"`
	Language    string                    `json:"language,omitempty"      desc:"Transcript language"`
	StartTimeMs string                    `json:"start_time_ms,omitempty" desc:"Transcript item start offset in milliseconds"`
	EndTimeMs   string                    `json:"end_time_ms,omitempty"   desc:"Transcript item end offset in milliseconds"`
	SentenceID  string                    `json:"sentence_id,omitempty"   desc:"Transcript sentence ID"`
}

// VCRecordingSpeakerOutput is the speaker metadata attached to transcript items.
type VCRecordingSpeakerOutput struct {
	ID       string `json:"id,omitempty"        desc:"Speaker user ID"`
	UserType int    `json:"user_type,omitempty" desc:"Speaker user type"`
	UserRole int    `json:"user_role,omitempty" desc:"Speaker role"`
	UserName string `json:"user_name,omitempty" desc:"Speaker display name"`
}

// VCRecordingStartedOutput is the flattened shape for vc.recording.recording_started_v1.
type VCRecordingStartedOutput struct {
	Type          string   `json:"type"                    desc:"Event type; always vc.recording.recording_started_v1"`
	EventID       string   `json:"event_id,omitempty"      desc:"Globally unique event ID; safe for deduplication"`
	Timestamp     string   `json:"timestamp,omitempty"     desc:"Event delivery time (ms timestamp string); taken from header.create_time when present" kind:"timestamp_ms"`
	RecordingID   string   `json:"recording_id,omitempty"  desc:"Recording ID"`
	Source        string   `json:"source,omitempty"        desc:"Recording source. Currently recording_bean only; consumers must branch by source for future sources."`
	SubscriberIDs []string `json:"subscriber_ids,omitempty" desc:"Subscribers receiving this user-scoped event"`
}

// VCRecordingTranscriptGeneratedOutput is the flattened shape for vc.recording.recording_transcript_generated_v1.
type VCRecordingTranscriptGeneratedOutput struct {
	Type            string                            `json:"type"                       desc:"Event type; always vc.recording.recording_transcript_generated_v1"`
	EventID         string                            `json:"event_id,omitempty"         desc:"Globally unique event ID; safe for deduplication"`
	Timestamp       string                            `json:"timestamp,omitempty"        desc:"Event delivery time (ms timestamp string); taken from header.create_time when present" kind:"timestamp_ms"`
	RecordingID     string                            `json:"recording_id,omitempty"     desc:"Recording ID"`
	Source          string                            `json:"source,omitempty"           desc:"Recording source. Currently recording_bean only; consumers must branch by source for future sources."`
	SubscriberIDs   []string                          `json:"subscriber_ids,omitempty"   desc:"Subscribers receiving this user-scoped event"`
	TranscriptItems []VCRecordingTranscriptItemOutput `json:"transcript_items,omitempty" desc:"Generated transcript items"`
}

// VCRecordingEndedOutput is the flattened shape for vc.recording.recording_ended_v1.
type VCRecordingEndedOutput struct {
	Type          string   `json:"type"                    desc:"Event type; always vc.recording.recording_ended_v1"`
	EventID       string   `json:"event_id,omitempty"      desc:"Globally unique event ID; safe for deduplication"`
	Timestamp     string   `json:"timestamp,omitempty"     desc:"Event delivery time (ms timestamp string); taken from header.create_time when present" kind:"timestamp_ms"`
	RecordingID   string   `json:"recording_id,omitempty"  desc:"Recording ID"`
	Source        string   `json:"source,omitempty"        desc:"Recording source. Currently recording_bean only; consumers must branch by source for future sources."`
	SubscriberIDs []string `json:"subscriber_ids,omitempty" desc:"Subscribers receiving this user-scoped event"`
	ObjectType    string   `json:"object_type,omitempty"   desc:"Artifact object type. Current value is minute; this is unstable and consumers should branch by object_type."`
	ObjectID      string   `json:"object_id,omitempty"     desc:"Artifact object ID. Current minute object_id is a Minutes token; avoid depending on this unless object_type is minute."`
}

type vcRecordingEnvelope struct {
	Header struct {
		EventID    string `json:"event_id"`
		EventType  string `json:"event_type"`
		CreateTime string `json:"create_time"`
	} `json:"header"`
	Event vcRecordingEvent `json:"event"`
}

type vcRecordingEvent struct {
	RecordingID     string                    `json:"recording_id"`
	Source          string                    `json:"source"`
	SubscriberIDs   []vcRecordingString       `json:"subscriber_ids"`
	TranscriptItems []vcRecordingTranscriptIn `json:"transcript_items"`
	ObjectType      string                    `json:"object_type"`
	ObjectID        string                    `json:"object_id"`
}

type vcRecordingTranscriptIn struct {
	Speaker     *vcRecordingSpeakerIn `json:"speaker"`
	Text        string                `json:"text"`
	Language    string                `json:"language"`
	StartTimeMs vcRecordingString     `json:"start_time_ms"`
	EndTimeMs   vcRecordingString     `json:"end_time_ms"`
	SentenceID  vcRecordingString     `json:"sentence_id"`
}

type vcRecordingSpeakerIn struct {
	ID       vcRecordingString `json:"id"`
	UserType int               `json:"user_type"`
	UserRole int               `json:"user_role"`
	UserName string            `json:"user_name"`
}

type vcRecordingString string

func processVCRecordingStarted(_ context.Context, _ event.APIClient, raw *event.RawEvent, _ map[string]string) (json.RawMessage, error) {
	envelope, ok := parseVCRecordingEnvelope(raw)
	if !ok {
		return raw.Payload, nil
	}
	out := &VCRecordingStartedOutput{
		Type:          recordingEventType(envelope, raw),
		EventID:       envelope.Header.EventID,
		Timestamp:     envelope.Header.CreateTime,
		RecordingID:   envelope.Event.RecordingID,
		Source:        envelope.Event.Source,
		SubscriberIDs: recordingStrings(envelope.Event.SubscriberIDs),
	}
	return json.Marshal(out)
}

func processVCRecordingTranscriptGenerated(_ context.Context, _ event.APIClient, raw *event.RawEvent, _ map[string]string) (json.RawMessage, error) {
	envelope, ok := parseVCRecordingEnvelope(raw)
	if !ok {
		return raw.Payload, nil
	}
	out := &VCRecordingTranscriptGeneratedOutput{
		Type:            recordingEventType(envelope, raw),
		EventID:         envelope.Header.EventID,
		Timestamp:       envelope.Header.CreateTime,
		RecordingID:     envelope.Event.RecordingID,
		Source:          envelope.Event.Source,
		SubscriberIDs:   recordingStrings(envelope.Event.SubscriberIDs),
		TranscriptItems: recordingTranscriptItems(envelope.Event.TranscriptItems),
	}
	return json.Marshal(out)
}

func processVCRecordingEnded(_ context.Context, _ event.APIClient, raw *event.RawEvent, _ map[string]string) (json.RawMessage, error) {
	envelope, ok := parseVCRecordingEnvelope(raw)
	if !ok {
		return raw.Payload, nil
	}
	out := &VCRecordingEndedOutput{
		Type:          recordingEventType(envelope, raw),
		EventID:       envelope.Header.EventID,
		Timestamp:     envelope.Header.CreateTime,
		RecordingID:   envelope.Event.RecordingID,
		Source:        envelope.Event.Source,
		SubscriberIDs: recordingStrings(envelope.Event.SubscriberIDs),
		ObjectType:    envelope.Event.ObjectType,
		ObjectID:      envelope.Event.ObjectID,
	}
	return json.Marshal(out)
}

func parseVCRecordingEnvelope(raw *event.RawEvent) (*vcRecordingEnvelope, bool) {
	var envelope vcRecordingEnvelope
	if err := json.Unmarshal(raw.Payload, &envelope); err != nil {
		return nil, false
	}
	return &envelope, true
}

func recordingEventType(envelope *vcRecordingEnvelope, raw *event.RawEvent) string {
	if envelope != nil && envelope.Header.EventType != "" {
		return envelope.Header.EventType
	}
	return raw.EventType
}

func recordingTranscriptItems(items []vcRecordingTranscriptIn) []VCRecordingTranscriptItemOutput {
	if len(items) == 0 {
		return nil
	}
	out := make([]VCRecordingTranscriptItemOutput, 0, len(items))
	for _, item := range items {
		out = append(out, VCRecordingTranscriptItemOutput{
			Speaker:     recordingSpeaker(item.Speaker),
			Text:        item.Text,
			Language:    item.Language,
			StartTimeMs: item.StartTimeMs.String(),
			EndTimeMs:   item.EndTimeMs.String(),
			SentenceID:  item.SentenceID.String(),
		})
	}
	return out
}

func recordingSpeaker(speaker *vcRecordingSpeakerIn) *VCRecordingSpeakerOutput {
	if speaker == nil {
		return nil
	}
	return &VCRecordingSpeakerOutput{
		ID:       speaker.ID.String(),
		UserType: speaker.UserType,
		UserRole: speaker.UserRole,
		UserName: speaker.UserName,
	}
}

func (s *vcRecordingString) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = vcRecordingString(str)
		return nil
	}
	var num json.Number
	if err := json.Unmarshal(data, &num); err != nil {
		return err
	}
	*s = vcRecordingString(num.String())
	return nil
}

func (s vcRecordingString) String() string {
	return string(s)
}

func recordingStrings(values []vcRecordingString) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.String())
	}
	return out
}
