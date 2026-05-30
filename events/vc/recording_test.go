// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package vc

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/larksuite/cli/internal/event"
)

func TestVCKeys_RecordingEventsRegistered(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())

	for _, tc := range []struct {
		eventType string
	}{
		{eventTypeRecordingStarted},
		{eventTypeRecordingTranscriptGenerated},
		{eventTypeRecordingEnded},
	} {
		t.Run(tc.eventType, func(t *testing.T) {
			def, ok := event.Lookup(tc.eventType)
			if !ok {
				t.Fatalf("%s should be registered via Keys()", tc.eventType)
			}
			if def.Schema.Custom == nil {
				t.Error("Processed key must set Schema.Custom")
			}
			if def.Schema.Native != nil {
				t.Error("Processed key must not set Schema.Native")
			}
			if def.Process == nil {
				t.Error("Process must not be nil for processed key")
			}
			if def.PreConsume == nil {
				t.Error("PreConsume must not be nil for processed key")
			}
			if len(def.Scopes) != 1 || def.Scopes[0] != "vc:recording:read" {
				t.Errorf("Scopes = %v", def.Scopes)
			}
			if len(def.AuthTypes) != 1 || def.AuthTypes[0] != "user" {
				t.Errorf("AuthTypes = %v", def.AuthTypes)
			}
			if len(def.RequiredConsoleEvents) != 1 || def.RequiredConsoleEvents[0] != tc.eventType {
				t.Errorf("RequiredConsoleEvents = %v", def.RequiredConsoleEvents)
			}
			if !strings.Contains(def.Description, "source") || !strings.Contains(def.Description, "recording_bean") {
				t.Errorf("Description should document source compatibility risk, got %q", def.Description)
			}
			if tc.eventType == eventTypeRecordingEnded && (!strings.Contains(def.Description, "object_type") || !strings.Contains(def.Description, "minute")) {
				t.Errorf("ended Description should document object_type/minute instability, got %q", def.Description)
			}
		})
	}
}

func TestProcessVCRecordingStarted(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())

	out := runRecordingProcess[VCRecordingStartedOutput](t, eventTypeRecordingStarted, processVCRecordingStarted, `{
		"schema": "2.0",
		"header": {
			"event_id": "ev_rec_start_001",
			"event_type": "vc.recording.recording_started_v1",
			"create_time": "1761782400000"
		},
		"event": {
			"recording_id": "recording_001",
			"source": "recording_bean",
			"subscriber_ids": [111, 222]
		}
	}`)

	if out.Type != eventTypeRecordingStarted {
		t.Errorf("Type = %q", out.Type)
	}
	if out.EventID != "ev_rec_start_001" || out.Timestamp != "1761782400000" {
		t.Errorf("EventID/Timestamp = %q/%q", out.EventID, out.Timestamp)
	}
	if out.RecordingID != "recording_001" || out.Source != "recording_bean" {
		t.Errorf("RecordingID/Source = %q/%q", out.RecordingID, out.Source)
	}
	assertStringSlice(t, out.SubscriberIDs, []string{"111", "222"})
}

func TestProcessVCRecordingTranscriptGenerated(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())

	out := runRecordingProcess[VCRecordingTranscriptGeneratedOutput](t, eventTypeRecordingTranscriptGenerated, processVCRecordingTranscriptGenerated, `{
		"schema": "2.0",
		"header": {
			"event_id": "ev_rec_transcript_001",
			"event_type": "vc.recording.recording_transcript_generated_v1",
			"create_time": "1761782400100"
		},
		"event": {
			"recording_id": "recording_001",
			"source": "recording_bean",
			"subscriber_ids": [111],
			"transcript_items": [
				{
					"speaker": {
						"id": "333",
						"user_type": 1,
						"user_role": 2,
						"user_name": "Alice"
					},
					"text": "hello world",
					"language": "en_us",
					"start_time_ms": "1200",
					"end_time_ms": "2400",
					"sentence_id": "987654321"
				}
			]
		}
	}`)

	if out.Type != eventTypeRecordingTranscriptGenerated {
		t.Errorf("Type = %q", out.Type)
	}
	if out.RecordingID != "recording_001" || out.Source != "recording_bean" {
		t.Errorf("RecordingID/Source = %q/%q", out.RecordingID, out.Source)
	}
	assertStringSlice(t, out.SubscriberIDs, []string{"111"})
	if len(out.TranscriptItems) != 1 {
		t.Fatalf("TranscriptItems len = %d, want 1", len(out.TranscriptItems))
	}
	item := out.TranscriptItems[0]
	if item.Text != "hello world" || item.Language != "en_us" {
		t.Errorf("Transcript text/language = %q/%q", item.Text, item.Language)
	}
	if item.StartTimeMs != "1200" || item.EndTimeMs != "2400" || item.SentenceID != "987654321" {
		t.Errorf("Transcript timing/id = %q/%q/%q", item.StartTimeMs, item.EndTimeMs, item.SentenceID)
	}
	if item.Speaker == nil {
		t.Fatal("Speaker should not be nil")
	}
	if item.Speaker.ID != "333" || item.Speaker.UserType != 1 || item.Speaker.UserRole != 2 || item.Speaker.UserName != "Alice" {
		t.Errorf("Speaker = %+v", item.Speaker)
	}
}

func TestProcessVCRecordingEnded(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())

	out := runRecordingProcess[VCRecordingEndedOutput](t, eventTypeRecordingEnded, processVCRecordingEnded, `{
		"schema": "2.0",
		"header": {
			"event_id": "ev_rec_end_001",
			"event_type": "vc.recording.recording_ended_v1",
			"create_time": "1761782400200"
		},
		"event": {
			"recording_id": "recording_001",
			"source": "recording_bean",
			"subscriber_ids": [111],
			"object_type": "minute",
			"object_id": "minute_token_001"
		}
	}`)

	if out.Type != eventTypeRecordingEnded {
		t.Errorf("Type = %q", out.Type)
	}
	if out.RecordingID != "recording_001" || out.Source != "recording_bean" {
		t.Errorf("RecordingID/Source = %q/%q", out.RecordingID, out.Source)
	}
	assertStringSlice(t, out.SubscriberIDs, []string{"111"})
	if out.ObjectType != "minute" || out.ObjectID != "minute_token_001" {
		t.Errorf("ObjectType/ObjectID = %q/%q", out.ObjectType, out.ObjectID)
	}
}

func TestProcessVCRecording_MalformedPayloadPassthrough(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())

	for _, tc := range []struct {
		name      string
		eventType string
		process   event.ProcessFunc
	}{
		{name: "started", eventType: eventTypeRecordingStarted, process: processVCRecordingStarted},
		{name: "transcript", eventType: eventTypeRecordingTranscriptGenerated, process: processVCRecordingTranscriptGenerated},
		{name: "ended", eventType: eventTypeRecordingEnded, process: processVCRecordingEnded},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := &event.RawEvent{
				EventType: tc.eventType,
				Payload:   json.RawMessage(`not json`),
				Timestamp: time.Now(),
			}
			got, err := tc.process(context.Background(), nil, raw, nil)
			if err != nil {
				t.Fatalf("Process should swallow parse errors, got %v", err)
			}
			if string(got) != "not json" {
				t.Errorf("malformed fallback output = %q, want original bytes", string(got))
			}
		})
	}
}

func TestVCRecording_PreConsumeSubscriptionLifecycle(t *testing.T) {
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())

	for _, tc := range []struct {
		eventType string
	}{
		{eventTypeRecordingStarted},
		{eventTypeRecordingTranscriptGenerated},
		{eventTypeRecordingEnded},
	} {
		t.Run(tc.eventType, func(t *testing.T) {
			def, ok := event.Lookup(tc.eventType)
			if !ok {
				t.Fatalf("%s should be registered via Keys()", tc.eventType)
			}

			type call struct {
				method string
				path   string
				body   any
			}
			var calls []call
			rt := &stubAPIClient{
				callFn: func(_ context.Context, method, path string, body any) (json.RawMessage, error) {
					calls = append(calls, call{method: method, path: path, body: body})
					return json.RawMessage(`{"code":0,"msg":"success","data":{}}`), nil
				},
			}

			cleanup, err := def.PreConsume(context.Background(), rt, nil)
			if err != nil {
				t.Fatalf("PreConsume error: %v", err)
			}
			if cleanup == nil {
				t.Fatal("cleanup must not be nil")
			}
			if len(calls) != 1 {
				t.Fatalf("calls after subscribe = %d, want 1", len(calls))
			}
			if calls[0].method != "POST" || calls[0].path != pathRecordingSubscribe {
				t.Fatalf("subscribe call = %+v", calls[0])
			}
			assertSubscriptionRequest(t, calls[0].body, tc.eventType)

			cleanup()
			if len(calls) != 2 {
				t.Fatalf("calls after cleanup = %d, want 2", len(calls))
			}
			if calls[1].method != "POST" || calls[1].path != pathRecordingUnsubscribe {
				t.Fatalf("unsubscribe call = %+v", calls[1])
			}
			assertSubscriptionRequest(t, calls[1].body, tc.eventType)
		})
	}
}

func runRecordingProcess[T any](t *testing.T, eventType string, process event.ProcessFunc, payload string) T {
	t.Helper()
	raw := &event.RawEvent{
		EventType: eventType,
		Payload:   json.RawMessage(payload),
		Timestamp: time.Now(),
	}
	got, err := process(context.Background(), nil, raw, nil)
	if err != nil {
		t.Fatalf("Process error: %v", err)
	}
	var out T
	if err := json.Unmarshal(got, &out); err != nil {
		t.Fatalf("Process output is not valid JSON: %v\nraw=%s", err, string(got))
	}
	return out
}

func assertStringSlice(t *testing.T, got, want []string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("slice len = %d, want %d; got=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice[%d] = %q, want %q; got=%v", i, got[i], want[i], got)
		}
	}
}
