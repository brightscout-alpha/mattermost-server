// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package slashcommands

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/mattermost/mattermost-server/v5/app"
	"github.com/mattermost/mattermost-server/v5/model"
	"github.com/mattermost/mattermost-server/v5/shared/i18n"
	"github.com/mattermost/mattermost-server/v5/shared/mlog"
)

type CustomStatusProvider struct {
}

const (
	CmdCustomStatus      = app.CmdCustomStatusTrigger
	CmdCustomStatusClear = "clear"

	DefaultCustomStatusEmoji = "speech_balloon"
)

func init() {
	app.RegisterCommandProvider(&CustomStatusProvider{})
}

func (*CustomStatusProvider) GetTrigger() string {
	return CmdCustomStatus
}

func (*CustomStatusProvider) GetCommand(a *app.App, T i18n.TranslateFunc) *model.Command {
	return &model.Command{
		Trigger:          CmdCustomStatus,
		AutoComplete:     true,
		AutoCompleteDesc: T("api.command_custom_status.desc"),
		AutoCompleteHint: T("api.command_custom_status.hint"),
		DisplayName:      T("api.command_custom_status.name"),
	}
}

func (*CustomStatusProvider) DoCommand(a *app.App, args *model.CommandArgs, message string) *model.CommandResponse {
	if !*a.Config().TeamSettings.EnableCustomUserStatuses {
		return nil
	}

	message = strings.TrimSpace(message)
	if message == CmdCustomStatusClear {
		if err := a.RemoveCustomStatus(args.UserId); err != nil {
			mlog.Error(err.Error())
			return &model.CommandResponse{Text: args.T("api.command_custom_status.clear.app_error"), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
		}

		return &model.CommandResponse{
			ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
			Text:         args.T("api.command_custom_status.clear.success"),
		}
	}

	customStatus := GetCustomStatus(message)
	if err := a.SetCustomStatus(args.UserId, customStatus); err != nil {
		mlog.Error(err.Error())
		return &model.CommandResponse{Text: args.T("api.command_custom_status.app_error"), ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL}
	}

	return &model.CommandResponse{
		ResponseType: model.COMMAND_RESPONSE_TYPE_EPHEMERAL,
		Text: args.T("api.command_custom_status.success", map[string]interface{}{
			"EmojiName":     ":" + customStatus.Emoji + ":",
			"StatusMessage": customStatus.Text,
		}),
	}
}

func GetCustomStatus(message string) *model.CustomStatus {
	customStatus := &model.CustomStatus{
		Emoji: DefaultCustomStatusEmoji,
		Text:  message,
	}

	firstEmojiLocations := model.ALL_EMOJI_PATTERN.FindIndex([]byte(message))
	if len(firstEmojiLocations) > 0 && firstEmojiLocations[0] == 0 {
		// emoji found at starting index
		customStatus.Emoji = message[firstEmojiLocations[0]+1 : firstEmojiLocations[1]-1]
		customStatus.Text = strings.TrimSpace(message[firstEmojiLocations[1]:])
	} else if message != "" {
		spaceSeparatedMessage := strings.Fields(message)
		if len(spaceSeparatedMessage) > 0 {
			emojiString := spaceSeparatedMessage[0]
			var unicode []string
			for utf8.RuneCountInString(emojiString) >= 1 {
				codepoint, size := utf8.DecodeRuneInString(emojiString)
				code := model.RuneToHexadecimalString(codepoint)
				unicode = append(unicode, code)
				emojiString = emojiString[size:]
			}

			unicodeString := removeUnicodeSkinTone(strings.Join(unicode, "-"))
			emoji, found := model.GetEmojiNameFromUnicode(unicodeString)
			if found {
				customStatus.Emoji = emoji
				textString := strings.Join(spaceSeparatedMessage[1:], " ")
				customStatus.Text = strings.TrimSpace(textString)
			}
		}
	}

	customStatus.TrimMessage()
	return customStatus
}

func removeUnicodeSkinTone(unicodeString string) string {
	skinToneDetectorRegex := regexp.MustCompile("-(1f3fb|1f3fc|1f3fd|1f3fe|1f3ff)")
	skinToneLocations := skinToneDetectorRegex.FindIndex([]byte(unicodeString))

	if len(skinToneLocations) > 0 {
		unicodeWithRemovedSkinTone := unicodeString[:skinToneLocations[0]] + unicodeString[skinToneLocations[1]:]
		unicodeWithVariationSelector := unicodeString[:skinToneLocations[0]] + "-fe0f" + unicodeString[skinToneLocations[1]:]
		if _, found := model.GetEmojiNameFromUnicode(unicodeWithRemovedSkinTone); found {
			unicodeString = unicodeWithRemovedSkinTone
		} else if _, found := model.GetEmojiNameFromUnicode(unicodeWithVariationSelector); found {
			unicodeString = unicodeWithVariationSelector
		}
	}

	return unicodeString
}
