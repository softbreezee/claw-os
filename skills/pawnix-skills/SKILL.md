---
name: pawnix-skills
description: How to create, structure, and manage skills in Pawnix. Load whenever the user asks about skills, wants to create a new skill, or needs to understand how the skill system works.
tags: [platform, guide, design]
metadata:
  pawnix:
    emoji: "🛠️"
    always: false
---

# Pawnix Skills System

Skills are the Pawnix equivalent of "apps". Each skill teaches an agent a
domain-specific playbook that it loads into its system prompt when relevant.

---

## Where Skills Live

Skills are discovered from multiple directories. Higher layers override lower
ones when the same name appears in multiple locations:

```
Layer 1 (highest): Agent workspace   <workspace>/skills/
Layer 2:           Per-agent         ~/.pawnix/agents/<id>/agent/skills/
Layer 3:           User custom       ~/.pawnix/skills/
Layer 4 (lowest):  Built-in          <project>/skills/
```

**Built-in** skills ship with the binary. They are read-only and always load
for all agents. Currently: `pawnix-skills`, `pawnix-skill-learner`, `debugging`,
`db-schema-designer`.

**Custom** skills live in `~/.pawnix/skills/`. They can be assigned to specific
agents or made common (available to all).

---

## Skill Formats

### Directory format (canonical)
```
my-skill/
├── SKILL.md         # Required: main content with YAML frontmatter
├── references/      # Optional: referenced docs
└── templates/       # Optional: output templates
```

### Single-file format (compact)
```
my-skill.md          # Filename = skill name, content = SKILL.md
```

---

## SKILL.md Structure

```yaml
---
name: my-skill
description: One-line summary. Include trigger contexts —
             "Use when the user asks about X, mentions Y, or needs Z"
metadata:
  pawnix:
    emoji: "🛠️"
    always: false       # true = full content always in system prompt
    os: [darwin]        # optional OS restriction
    requires:
      bins: [git]       # all must exist on PATH
      anyBins: [python3, node]  # at least one
      env: [API_KEY]    # required env vars
    primaryEnv: API_KEY # maps config apiKey to this env
---

# Skill Title

Step-by-step instructions in markdown...
```

### Description Tips

The description determines when agents load the skill. Make it:
- Specific: include what the skill does AND when it should trigger
- Pushy but not spammy: "Use when the user asks about X, mentions Y, or needs Z"

Bad: `"Format data"`  
Good: `"Format CSV data into clean tables. Use when the user mentions spreadsheets, data formatting, column alignment, or CSV cleanup."`

---

## Agent Assignment (Multi-to-Multi)

A skill can be assigned to multiple agents. This is configured in
`~/.pawnix/pawnix.json`, NOT by copying files:

```json
{
  "skills": {
    "entries": {
      "buffett":  { "enabled": true, "agents": ["trend-trader", "deep-research"] },
      "munger":   { "enabled": true, "agents": ["trend-trader"] },
      "pawnix-ashare-data": { "enabled": true, "agents": [] }
    }
  }
}
```

**Rules:**
- `agents: []` — common, all agents load it
- `agents: ["trend-trader"]` — only that agent loads it
- `agents: ["trend-trader", "deep-research"]` — both agents load it
- A skill not in entries defaults to common
- Built-in skills always load, regardless of agents config

**Runtime calculation for an agent:**
```
load = all built-in + all common + all agents that include this agent
```

---

## Creating a Skill

1. Create the directory: `mkdir -p ~/.pawnix/skills/my-skill`
2. Write `SKILL.md` with frontmatter + body
3. Optionally add `references/` for large supporting docs
4. The skill is auto-discovered on next agent message
5. Assign agents: go to Skills page in Web UI, use the `[+]` dropdown

---

## Writing Guidelines

- Use imperative form: "Run the command", not "You should run the command"
- Explain **why**, not just what
- Include examples for output formats
- Use `{baseDir}` to reference files within the skill directory
- Keep SKILL.md under 500 lines — move large content to `references/`
- Split into multiple skills if a skill covers too many topics
