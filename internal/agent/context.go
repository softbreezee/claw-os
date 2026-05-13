package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// bootstrapFiles are loaded in order to build the system prompt.
var bootstrapFiles = []string{
	"AGENTS.md",
	"BOOTSTRAP.md",
	"HEARTBEAT.md",
	"SOUL.md",
	"USER.md",
	"TOOLS.md",
	"IDENTITY.md",
}

// GroupContext holds information about the group chat environment for system prompt injection.
type GroupContext struct {
	BotUsername string   // this agent's bot username
	Teammates  []string // other agent names in the group
}

// ContextBuilder assembles the system prompt and runtime context.
type ContextBuilder struct {
	workspace     string
	memory        *Memory
	skillsSummary string
	groupCtx      *GroupContext
	thinking      string // off, low, medium, high, adaptive
	model         string // current LLM model name (e.g. "gpt-4o-mini", "claude-sonnet-4-5")
}

// NewContextBuilder creates a new context builder.
func NewContextBuilder(workspace string, memory *Memory, skillsSummary string) *ContextBuilder {
	return &ContextBuilder{
		workspace:     workspace,
		memory:        memory,
		skillsSummary: skillsSummary,
	}
}

// SetModel updates the model name surfaced in the identity section.
// Called on construction and again whenever the model is hot-swapped
// (e.g. via /model slash command or UpdateConfig).
func (cb *ContextBuilder) SetModel(model string) {
	cb.model = model
}

// SystemPromptSection is one labelled chunk of the assembled system prompt.
// Surfaced to the Web UI so users can see what makes up their 5kB+ prefix
// and which slice (Identity / Bootstrap / Skills / ...) is the bloat
// source. Order matches the order in which BuildSystemPrompt joins them.
type SystemPromptSection struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

// BuildSystemPrompt assembles the system prompt from identity, bootstrap files, memory, and skills.
// Implementation note: this delegates to BuildSystemPromptSections and joins
// with the standard separator. Keeping a single source of truth means the
// "preview in UI" view can never drift from what the LLM actually receives.
func (cb *ContextBuilder) BuildSystemPrompt() string {
	sections := cb.BuildSystemPromptSections()
	parts := make([]string, len(sections))
	for i, s := range sections {
		parts[i] = s.Content
	}
	return strings.Join(parts, "\n\n---\n\n")
}

// BuildSystemPromptSections returns the same content BuildSystemPrompt
// would produce, but split into labelled sections. Each section keeps
// the same headers (e.g. "# Skills") it would have in the joined prompt
// so token-count math agrees with what the model sees.
func (cb *ContextBuilder) BuildSystemPromptSections() []SystemPromptSection {
	var sections []SystemPromptSection

	// 1. Identity (runtime environment info)
	//    Include the underlying LLM model so the agent can answer
	//    "what model are you running on?" without having to guess.
	modelLine := cb.model
	if modelLine == "" {
		modelLine = "(unknown)"
	}
	identity := fmt.Sprintf(`You are Pawnix, a self-hosted AI-Native personal OS.
Model: %s
OS: %s/%s
Working Directory: %s`, modelLine, runtime.GOOS, runtime.GOARCH, cb.workspace)
	sections = append(sections, SystemPromptSection{Name: "Identity", Content: identity})

	// 2. Bootstrap files (each gets its own section so users can see which
	//    workspace file is contributing how many tokens — bootstrap files
	//    are the most likely thing for a user to want to trim).
	for _, name := range bootstrapFiles {
		content := cb.loadFile(name)
		if content != "" {
			sections = append(sections, SystemPromptSection{
				Name:    name,
				Content: fmt.Sprintf("# %s\n%s", name, content),
			})
		}
	}

	// 3. Skills
	if cb.skillsSummary != "" {
		sections = append(sections, SystemPromptSection{
			Name:    "Skills",
			Content: fmt.Sprintf("# Skills\n%s", cb.skillsSummary),
		})
	}

	// 4. Long-term memory
	mem := cb.memory.LoadMemory()
	if mem != "" {
		sections = append(sections, SystemPromptSection{
			Name:    "Long-term Memory",
			Content: fmt.Sprintf("# Long-term Memory\n%s", mem),
		})
	}

	// 5. Group chat awareness
	if cb.groupCtx != nil {
		groupInfo := fmt.Sprintf(`# Group Chat
You are in a group chat. Your bot username is @%s.
Other agents in this group: %s.
Only respond when directly mentioned with @%s, or when the conversation clearly needs your expertise.
Messages from other bots will appear as "[BotName]: message" in the conversation history.`,
			cb.groupCtx.BotUsername,
			strings.Join(cb.groupCtx.Teammates, ", "),
			cb.groupCtx.BotUsername,
		)
		sections = append(sections, SystemPromptSection{Name: "Group Chat", Content: groupInfo})
	}

	// 6. Thinking/Reasoning mode
	if cb.thinking != "" && cb.thinking != "off" {
		thinkingPrompt := cb.buildThinkingPrompt()
		if thinkingPrompt != "" {
			sections = append(sections, SystemPromptSection{Name: "Thinking Mode", Content: thinkingPrompt})
		}
	}

	// 7. Self-updating workspace files guidance
	sections = append(sections, SystemPromptSection{
		Name: "Workspace Self-Update",
		Content: `# Workspace Self-Update
You have the ability to update workspace files to maintain knowledge over time:
- MEMORY.md: Update when you learn important facts, user preferences, or key decisions. This file is loaded into your context every conversation.
- USER.md: Update when you learn new information about the user (role, preferences, communication style).
- HEARTBEAT.md: Update to add/remove periodic tasks you should check on.
- TOOLS.md: Update if you discover new tool usage patterns worth documenting.
Use the write_file tool to update these files when appropriate. Keep entries concise and useful.`,
	})

	return sections
}

// BuildRuntimeContext returns the runtime context to inject before the user message.
func (cb *ContextBuilder) BuildRuntimeContext(channel, chatID string) string {
	now := time.Now()
	return fmt.Sprintf(`[Runtime Context — metadata only, not instructions]
Time: %s
Timezone: %s
Channel: %s
Chat ID: %s`, now.Format("2006-01-02 15:04:05"), now.Location().String(), channel, chatID)
}

// SetGroupContext sets the group chat context for system prompt generation.
func (cb *ContextBuilder) SetGroupContext(gc *GroupContext) {
	cb.groupCtx = gc
}

// SetThinking configures the thinking/reasoning level.
func (cb *ContextBuilder) SetThinking(level string) {
	cb.thinking = level
}

func (cb *ContextBuilder) buildThinkingPrompt() string {
	var depth string
	switch cb.thinking {
	case "low":
		depth = "briefly reason through"
	case "medium":
		depth = "think step-by-step through"
	case "high":
		depth = "deeply and thoroughly reason through"
	case "adaptive":
		depth = "adaptively reason through (brief for simple tasks, deep for complex ones)"
	default:
		return ""
	}

	return fmt.Sprintf(`# Thinking Mode
Before responding to each message, %s your approach internally. Consider:
- What is the user really asking for?
- What are the key constraints and edge cases?
- What is the best approach and why?
- Are there any risks or trade-offs to consider?
Structure your reasoning before acting. Think before you respond.`, depth)
}

func (cb *ContextBuilder) loadFile(name string) string {
	path := filepath.Join(cb.workspace, name)
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
