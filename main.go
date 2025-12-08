package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// styles
var (
	appStyle      = lipgloss.NewStyle().Margin(1, 2)
	viewportStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("62")).
			Padding(1, 2)

	textareaStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("240")).
			Padding(1, 2)

	messageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205"))

	botMessageStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("86"))
)

type errMsg error
type cliResponseMsg string

type model struct {
	viewport   viewport.Model
	textarea   textarea.Model
	messages   []string
	cliLoading bool
	err        error
}

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Enter your message here"
	ta.Focus()
	ta.Prompt = "| "
	ta.CharLimit = 2000
	ta.SetWidth(30)
	ta.SetHeight(3)
	ta.ShowLineNumbers = true
	ta.KeyMap.InsertNewline.SetEnabled(true)

	vp := viewport.New(30, 5)
	vp.SetContent("Chat successfully initialized. Type a message below.")

	return model{
		viewport:   vp,
		textarea:   ta,
		messages:   []string{},
		cliLoading: false,
		err:        nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
	)

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+j", "shift+enter":
			// shift+enter 가 ctrl+j 로 들어옴
			m.textarea.SetValue(m.textarea.Value() + "\n")
		}
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyUp:
			m.viewport.ScrollUp(1)
		case tea.KeyDown:
			m.viewport.ScrollDown(1)
		case tea.KeyEnter:
			if m.cliLoading {
				return m, nil
			}

			userInput := m.textarea.Value()
			if userInput == "" {
				return m, nil
			}

			m.messages = append(m.messages, messageStyle.Render("User : ")+userInput)

			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			m.textarea.Reset()
			m.cliLoading = true

			return m, tea.Batch(tiCmd, runChatCommand(userInput))
		}
	case cliResponseMsg:
		m.cliLoading = false
		response := string(msg)

		m.messages = append(m.messages, botMessageStyle.Render("Bot : ")+response)
		m.messages = append(m.messages, "")

		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

		return m, tea.Batch(tiCmd, vpCmd)
	case tea.WindowSizeMsg:
		headerHeight := 0
		footerHeight := 6
		varticalMarginHeight := headerHeight + footerHeight

		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - varticalMarginHeight

		m.textarea.SetWidth(msg.Width - 4)

	case errMsg:
		m.err = msg
	}

	return m, tea.Batch(tiCmd, vpCmd)
}

func (m model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n", m.err)
	}

	// 뷰포트 렌더링 (스타일 적용)
	chatBox := viewportStyle.Render(m.viewport.View())

	// 입력창 렌더링
	inputBox := m.textarea.View()

	// 로딩 표시
	if m.cliLoading {
		inputBox = "Thinking..."
	}

	return appStyle.Render(fmt.Sprintf(
		"%s\n%s",
		chatBox,
		inputBox,
	))
}

// --- 6. 외부 명령 실행 함수 (Integration) ---
// 실제 ClaudeCode나 Gemini CLI를 여기서 호출합니다.
func runChatCommand(input string) tea.Cmd {
	return func() tea.Msg {
		// [실제 연동 방법]
		// 아래 exec.Command 부분을 실제 사용하려는 툴로 변경하세요.
		// 예: exec.Command("claude", "--message", input)
		// 예: exec.Command("gemini", "prompt", input)

		// 여기서는 테스트를 위해 'echo' 명령어로 시뮬레이션합니다.
		// 실제 AI 툴이 설치되어 있다면 주석을 해제하고 교체하세요.

		// cmd := exec.Command("claude", "p", input) // 예시
		cmd := exec.Command("echo", "Simulated AI Response to: "+input)

		out, err := cmd.CombinedOutput()
		if err != nil {
			return cliResponseMsg("Error executing command: " + err.Error())
		}

		return cliResponseMsg(string(out))
	}
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
