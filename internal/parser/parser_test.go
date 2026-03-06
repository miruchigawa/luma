package parser

import (
	"testing"
	"time"

	waProto "go.mau.fi/whatsmeow/proto/waE2E"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	"google.golang.org/protobuf/proto"
)

func TestParse(t *testing.T) {
	sender := types.JID{User: "123", Server: types.DefaultUserServer}
	chat := types.JID{User: "456", Server: types.DefaultUserServer}
	ts := time.Now()

	createEvt := func(msg *waProto.Message) *events.Message {
		return &events.Message{
			Info: types.MessageInfo{
				MessageSource: types.MessageSource{
					Sender:  sender,
					Chat:    chat,
					IsGroup: false,
				},
				ID:        "msg_id",
				PushName:  "TestUser",
				Timestamp: ts,
			},
			Message: msg,
		}
	}

	tests := []struct {
		name     string
		prefix   string
		msg      *waProto.Message
		validate func(*testing.T, *ParsedMessage)
	}{
		{
			name:   "plain text (not a command)",
			prefix: "!",
			msg: &waProto.Message{
				Conversation: proto.String("hello world"),
			},
			validate: func(t *testing.T, pm *ParsedMessage) {
				if pm.IsCommand {
					t.Errorf("Expected IsCommand to be false")
				}
				if pm.Body != "hello world" {
					t.Errorf("Expected Body 'hello world', got '%s'", pm.Body)
				}
			},
		},
		{
			name:   "command !ping",
			prefix: "!",
			msg: &waProto.Message{
				Conversation: proto.String("!ping"),
			},
			validate: func(t *testing.T, pm *ParsedMessage) {
				if !pm.IsCommand {
					t.Errorf("Expected IsCommand to be true")
				}
				if pm.CommandName != "ping" {
					t.Errorf("Expected CommandName 'ping', got '%s'", pm.CommandName)
				}
				if len(pm.Args) != 0 {
					t.Errorf("Expected 0 args, got %d", len(pm.Args))
				}
			},
		},
		{
			name:   "command with args !echo hello world",
			prefix: "!",
			msg: &waProto.Message{
				Conversation: proto.String("!echo hello world"),
			},
			validate: func(t *testing.T, pm *ParsedMessage) {
				if !pm.IsCommand {
					t.Errorf("Expected IsCommand to be true")
				}
				if pm.CommandName != "echo" {
					t.Errorf("Expected CommandName 'echo', got '%s'", pm.CommandName)
				}
				if pm.RawArgs != "hello world" {
					t.Errorf("Expected RawArgs 'hello world', got '%s'", pm.RawArgs)
				}
			},
		},
		{
			name:   "image message with caption command",
			prefix: "!",
			msg: &waProto.Message{
				ImageMessage: &waProto.ImageMessage{
					Caption: proto.String("!sticker"),
				},
			},
			validate: func(t *testing.T, pm *ParsedMessage) {
				if pm.MediaType != MediaImage {
					t.Errorf("Expected MediaType MediaImage, got %s", pm.MediaType)
				}
				if !pm.IsCommand {
					t.Errorf("Expected IsCommand to be true")
				}
				if pm.CommandName != "sticker" {
					t.Errorf("Expected CommandName 'sticker', got '%s'", pm.CommandName)
				}
			},
		},
		{
			name:   "reply to a message",
			prefix: "!",
			msg: &waProto.Message{
				ExtendedTextMessage: &waProto.ExtendedTextMessage{
					Text: proto.String("!quote"),
					ContextInfo: &waProto.ContextInfo{
						StanzaID:    proto.String("quoted_id"),
						Participant: proto.String("789@s.whatsapp.net"),
						QuotedMessage: &waProto.Message{
							Conversation: proto.String("original message"),
						},
					},
				},
			},
			validate: func(t *testing.T, pm *ParsedMessage) {
				if pm.QuotedMsg == nil {
					t.Fatalf("Expected QuotedMsg to not be nil")
				}
				if pm.QuotedMsg.Body != "original message" {
					t.Errorf("Expected quoted body 'original message', got '%s'", pm.QuotedMsg.Body)
				}
				if pm.QuotedMsg.ID != "quoted_id" {
					t.Errorf("Expected quoted ID 'quoted_id', got '%s'", pm.QuotedMsg.ID)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			evt := createEvt(tt.msg)
			pm := Parse(evt, []string{tt.prefix})
			tt.validate(t, pm)
		})
	}
}
