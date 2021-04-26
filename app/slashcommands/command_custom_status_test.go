// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package slashcommands

import (
	"testing"

	"github.com/mattermost/mattermost-server/v5/model"
)

func TestGetCustomStatus(t *testing.T) {
	for msg, expected := range map[string]model.CustomStatus{
		"":              model.CustomStatus{Emoji: DefaultCustomStatusEmoji, Text: ""},
		"hey":           model.CustomStatus{Emoji: DefaultCustomStatusEmoji, Text: "hey"},
		":cactus: hurt": model.CustomStatus{Emoji: "cactus", Text: "hurt"},
		"✋":             model.CustomStatus{Emoji: "raised_hand", Text: ""},
		"✋ handsup":     model.CustomStatus{Emoji: "raised_hand", Text: "handsup"},
		"👪 family":      model.CustomStatus{Emoji: "family", Text: "family"},
		"👙 swimming":    model.CustomStatus{Emoji: "bikini", Text: "swimming"},
	} {
		actual := GetCustomStatus(msg)
		if actual.Emoji != expected.Emoji || actual.Text != expected.Text {
			t.Errorf("expected `%v`, got `%v`", expected, *actual)
		}
	}
}
