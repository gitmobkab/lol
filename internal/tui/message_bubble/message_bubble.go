package message_bubble

import (
	"strings"

	"charm.land/glamour/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
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

	header := renderHeader(msg.Sender, headerStyle, innerWidth)
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

// renderHeader builds the header row with the sender name on the left and
// a copy-hint icon (⎘) on the right. Clicking the header row copies the message.
func renderHeader(sender string, style lipgloss.Style, innerWidth int) string {
	const icon = " ⎘"
	iconW := lipgloss.Width(icon)
	// headerStyle has PaddingLeft(1); total block width = innerWidth, so content = innerWidth-1.
	contentW := innerWidth - 1
	senderMaxW := max(contentW-iconW, 0)
	senderStr := ansi.Truncate(sender, senderMaxW, "")
	fill := strings.Repeat(" ", max(senderMaxW-lipgloss.Width(senderStr), 0))
	return style.Width(innerWidth).Render(senderStr + fill + icon)
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
