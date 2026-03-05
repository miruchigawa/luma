package parser

import (
	"strings"
	"time"

	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
)

type MediaType string

const (
	MediaNone     MediaType = "none"
	MediaImage    MediaType = "image"
	MediaVideo    MediaType = "video"
	MediaAudio    MediaType = "audio"
	MediaDocument MediaType = "document"
	MediaSticker  MediaType = "sticker"
)

type QuotedMessage struct {
	ID        string
	From      types.JID
	Body      string
	MediaType MediaType
}

// ParsedMessage unpacks the raw Protobuf message into a cleaner structure.
type ParsedMessage struct {
	ID          string
	From        types.JID
	Chat        types.JID
	IsGroup     bool
	SenderName  string
	Body        string
	IsCommand   bool
	CommandName string
	Args        []string
	RawArgs     string
	MediaType   MediaType
	QuotedMsg   *QuotedMessage
	Timestamp   time.Time
}

// Parse extracts relevant information from a raw WhatsApp message event.
//
// Key design decisions:
// - Centralizes message unpacking so commands don't deal with Protobuf internals.
// - Normalizes commands by stripping whitespace, lowercasing names.
// - Gracefully handles different message types (Text, Image+Caption, Video+Caption, etc.).
func Parse(evt *events.Message, prefix string) *ParsedMessage {
	if evt.Message == nil {
		return nil
	}

	pm := &ParsedMessage{
		ID:         evt.Info.ID,
		From:       evt.Info.Sender,
		Chat:       evt.Info.Chat,
		IsGroup:    evt.Info.IsGroup,
		SenderName: evt.Info.PushName,
		Timestamp:  evt.Info.Timestamp,
		MediaType:  MediaNone,
	}

	msg := evt.Message
	var body string

	if msg.GetConversation() != "" {
		body = msg.GetConversation()
	} else if msg.GetExtendedTextMessage() != nil {
		extt := msg.GetExtendedTextMessage()
		body = extt.GetText()

		if ctxInfo := extt.GetContextInfo(); ctxInfo != nil && ctxInfo.GetQuotedMessage() != nil {
			qMsg := ctxInfo.GetQuotedMessage()
			qBody := ""
			qMedia := MediaNone

			if qMsg.GetConversation() != "" {
				qBody = qMsg.GetConversation()
			} else if qMsg.GetExtendedTextMessage() != nil {
				qBody = qMsg.GetExtendedTextMessage().GetText()
			} else if qMsg.GetImageMessage() != nil {
				qBody = qMsg.GetImageMessage().GetCaption()
				qMedia = MediaImage
			} else if qMsg.GetVideoMessage() != nil {
				qBody = qMsg.GetVideoMessage().GetCaption()
				qMedia = MediaVideo
			} else if qMsg.GetAudioMessage() != nil {
				qMedia = MediaAudio
			} else if qMsg.GetDocumentMessage() != nil {
				qMedia = MediaDocument
			} else if qMsg.GetStickerMessage() != nil {
				qMedia = MediaSticker
			}

			qFrom, _ := types.ParseJID(ctxInfo.GetParticipant())

			pm.QuotedMsg = &QuotedMessage{
				ID:        ctxInfo.GetStanzaID(),
				From:      qFrom,
				Body:      qBody,
				MediaType: qMedia,
			}
		}
	} else if msg.GetImageMessage() != nil {
		body = msg.GetImageMessage().GetCaption()
		pm.MediaType = MediaImage
	} else if msg.GetVideoMessage() != nil {
		body = msg.GetVideoMessage().GetCaption()
		pm.MediaType = MediaVideo
	} else if msg.GetAudioMessage() != nil {
		pm.MediaType = MediaAudio
	} else if msg.GetDocumentMessage() != nil {
		pm.MediaType = MediaDocument
	} else if msg.GetStickerMessage() != nil {
		pm.MediaType = MediaSticker
	}

	pm.Body = strings.TrimSpace(body)

	if prefix != "" && strings.HasPrefix(pm.Body, prefix) {
		cmdLine := strings.TrimPrefix(pm.Body, prefix)
		cmdLine = strings.TrimSpace(cmdLine)
		if cmdLine != "" {
			pm.IsCommand = true
			parts := strings.Fields(cmdLine)
			pm.CommandName = strings.ToLower(parts[0])
			if len(parts) > 1 {
				pm.Args = parts[1:]

				firstArgIdx := strings.Index(cmdLine, parts[1])
				if firstArgIdx != -1 {
					pm.RawArgs = cmdLine[firstArgIdx:]
				} else {
					pm.RawArgs = strings.Join(pm.Args, " ")
				}
			}
		}
	}

	return pm
}
