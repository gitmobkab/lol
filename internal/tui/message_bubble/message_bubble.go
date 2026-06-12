package message_bubble

import (
	"strings"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
)

const WidthRatio = 55

// Render builds a styled terminal bubble for a chat message.
// It uses glamour to render the body as markdown.
func Render(msg Message, vpWidth int) string {
	maxWidth := max(vpWidth*WidthRatio/100, 20)
	maxWidth = min(maxWidth, vpWidth-2)
	innerWidth := max(maxWidth-2, 4) // subtract border (1 left + 1 right)

	var headerStyle, borderStyle lipgloss.Style
	switch {
	case msg.Kind == KindSystem:
		headerStyle = headerSystemStyle
		borderStyle = bubbleBorderStyle
	case msg.Kind == KindDM:
		headerStyle = headerDMStyle
		borderStyle = bubbleBorderDMStyle
	case msg.Own:
		headerStyle = headerOwnStyle
		borderStyle = bubbleBorderOwnStyle
	default:
		headerStyle = headerOtherStyle
		borderStyle = bubbleBorderStyle
	}

	bodyWidth := innerWidth - 2 // subtract body padding (PaddingLeft+PaddingRight = 2)
	normalized := strings.ReplaceAll(msg.Body, "\n", "\n\n")
	renderedBody := renderMarkdown(normalized, bodyWidth)

	header := headerStyle.Width(innerWidth).Render(msg.Sender)
	body := bubbleBodyStyle.Width(innerWidth).Render(renderedBody)
	timeRow := bubbleTimeStyle.Width(innerWidth).Align(lipgloss.Right).Render(msg.Ts)
	inner := lipgloss.JoinVertical(lipgloss.Left, header, body, timeRow)
	bubble := borderStyle.Render(inner)

	bw := lipgloss.Width(bubble)
	switch {
	case msg.Own:
		if pad := vpWidth - bw; pad > 0 {
			return indentLines(bubble, pad)
		}
	case msg.Kind == KindSystem:
		if pad := (vpWidth - bw) / 2; pad > 0 {
			return indentLines(bubble, pad)
		}
	}
	return bubble
}

func renderMarkdown(text string, width int) string {
	r, err := glamour.NewTermRenderer(
		glamour.WithWordWrap(width),
		glamour.WithStandardStyle(currentGlamourStyle),
	)
	if err != nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return strings.TrimSpace(out)
}

func indentLines(s string, n int) string {
	prefix := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = prefix + l
	}
	return strings.Join(lines, "\n")
}
