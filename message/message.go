/*
   Copyright 2013 Matthew Collins (purggames@gmail.com)

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package message

import (
	"bytes"
	"encoding/json"
)

type Message struct {
	//Plain text. Don't use Text and Translate at the same time
	Text string `json:"text,omitempty"`
	//Translatable string. Don't use Text and Translate at the same time
	Translate string `json:"translate,omitempty"`
	//Arguments for Translate. Replaces %s in the translated string
	With []*Message `json:"with,omitempty"`
	//Messages in this slice will be appened to the final string
	Extra []*Message `json:"extra,omitempty"`
	//Controls whether the text is bold
	Bold *bool `json:"bold,omitempty"`
	//Controls whether the text is italic
	Italic *bool `json:"italic,omitempty"`
	//Controls whether the text is underlined
	Underlined *bool `json:"underlined,omitempty"`
	//Controls whether the text is striked out
	Strikethrough *bool `json:"strikethrough,omitempty"`
	//Controls whether the text is randomised
	Obfuscated *bool `json:"obfuscated,omitempty"`
	//Controls the color of the text
	Color Color `json:"color,omitempty"`
	//Controls what happens if the text is clicked
	ClickEvent *ClickEvent `json:"clickEvent,omitempty"`
	//Controls what happens if the text is hovered over
	HoverEvent *HoverEvent `json:"hoverEvent,omitempty"`
}

var (
	_false = false
	False  = &_false
	_true  = true
	True   = &_true
)

//Returns a json encoded string of the message
func (m *Message) JSONString() string {
	res, _ := json.Marshal(m)
	return string(res)
}

//Returns a plain string of the message
func (m *Message) String() string {
	var buf bytes.Buffer
	m.string(&buf)
	return buf.String()
}

func (m *Message) string(buf *bytes.Buffer) {
	if m.Text != "" {
		buf.WriteString(m.Text)
	} else {
		//TODO
		panic("Translatable strings cannot be stringified yet")
	}
	if m.Extra != nil {
		for _, e := range m.Extra {
			e.string(buf)
		}
	}
}

type ClickEvent struct {
	Action ClickAction `json:"action"`
	Value  string      `json:"value"`
}

type HoverEvent struct {
	Action HoverAction `json:"action"`
	Value  *Message    `json:"value"`
}

//A minecraft text click event action
type ClickAction string

//Click events
const (
	//Opens the url in Value
	OpenUrl ClickAction = "open_url"
	//Opens the file at Value
	OpenFile ClickAction = "open_file"
	//Runs the chat command in Value
	RunCommand ClickAction = "run_command"
	//Places the chat command in Value into the player's chat box
	SuggestCommand ClickAction = "suggest_command"
)

//A minecraft text click event action
type HoverAction string

//Hover events
const (
	//Shows the message at Value on hover
	ShowText HoverAction = "show_text"
	//Shows the achievement at Value on hover
	ShowAchievement HoverAction = "show_achievement"
	//Shows the item at Value on hover
	ShowItem HoverAction = "show_item"
)

//A minecraft chat color
type Color string

//Supported colors
const (
	Black       Color = "black"
	DarkBlue    Color = "dark_blue"
	DarkGreen   Color = "dark_green"
	DarkAqua    Color = "dark_aqua"
	DarkRed     Color = "dark_red"
	DarkPurple  Color = "dark_purple"
	Gold        Color = "gold"
	Gray        Color = "gray"
	DarkGray    Color = "dark_gray"
	Blue        Color = "blue"
	Green       Color = "green"
	Aqua        Color = "aqua"
	Red         Color = "red"
	LightPurple Color = "red"
	Yellow      Color = "yellow"
	White       Color = "white"
)
