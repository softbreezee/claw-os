# Multimodal Attachments

## Scope

Stage 1 ships **image attachments via the Web UI**. PDFs, audio, and
multi-channel attachment ingestion are explicitly out of scope for now
(see "Future stages" below).

## End-to-end flow

```
Web UI (paperclip / paste / drop)
    │ multipart/form-data
    ▼
POST /api/chat/submit
    │ - parse form
    │ - persist each file to ~/.pawnix/uploads/<sha256>.<ext>
    │ - build []bus.Attachment
    ▼
taskrunner.SubmitWithOptions(..., Attachments)
    │ stash in pendingAttachments[taskID] (in-memory, per-process)
    ▼
runner.run() pops attachments → ContextWithAttachments(ctx, ...)
    ▼
agent.HandleWebChatStream  reads atts from ctx, copies into bus.InboundMessage
    ▼
agent.HandleMessage  → buildUserMessage(msg)
    │ images → buildContentParts → base64 data: URLs
    ▼
provider.OpenAI / provider.Anthropic
    serialise ContentParts to upstream wire format
```

## Key design choices

| Decision | Choice | Why |
|---|---|---|
| File storage | Local content-addressed (`sha256.ext`) | Free dedup, no external deps |
| Wire format | base64 inline `data:` URL | Hosted LLM APIs can't fetch from `localhost`; inlining sidesteps reachability |
| Persistence | Session stores **paths**, not bytes | Avoids MB-sized base64 in PG; bytes re-read on demand |
| Attachment plumbing | per-task in-memory map (`pendingAttachments`) | No DB schema change; mirrors `pendingModels` |
| Non-image fallback | Synthetic `[attached: name.ext]` text breadcrumb | Model knows something was sent; can call a tool later |

## Components added

| File | Role |
|---|---|
| [`internal/upload/store.go`](../internal/upload/store.go) | Content-addressed local file store |
| [`internal/agent/attachments.go`](../internal/agent/attachments.go) | `buildUserMessage` + `buildContentParts` + base64 encoder |
| [`internal/agent/events.go`](../internal/agent/events.go) | `ContextWithAttachments` / `AttachmentsFromContext` |
| [`internal/bus/bus.go`](../internal/bus/bus.go) | `Attachment` type + `InboundMessage.Attachments` |
| [`internal/taskrunner/runner.go`](../internal/taskrunner/runner.go) | `SubmitOptions{Attachments}` + per-task plumbing |
| [`internal/setup/handlers_chattasks.go`](../internal/setup/handlers_chattasks.go) | multipart/form-data branch in `handleChatSubmit` |
| [`internal/setup/handlers_files.go`](../internal/setup/handlers_files.go) | `GET /api/files` (workspace + uploads), `POST /api/workspace/open` |
| [`web/src/app/chat/page.tsx`](../web/src/app/chat/page.tsx) | Paperclip / paste / drop / preview chips, in-bubble image rendering, lightbox, "Open folder" button |

## Limits

- **Per file**: 25 MiB (constant `maxAttachmentBytes`); enforced both in JS and Go.
- **Per request**: bounded by `maxMultipartMemory` (32 MiB in RAM, larger spills to `/tmp`).
- **Cleanup**: none yet — the `~/.pawnix/uploads/` dir grows monotonically.

## Future stages

- **Stage 2** — PDF support: extend `ContentPart` with `Document`; route by
  per-model capability declared in `ModelEntry.Input`. Anthropic native
  document blocks for Claude 3.5+; `pdftotext` fallback otherwise.
- **Stage 3** — Cross-channel ingest: extract attachments from Discord
  (`m.Attachments`) and Slack (`ev.Files`), promote `PhotoURL` consumers
  to `Attachments` in Telegram.
- **Stage 4** — Storage hygiene: configurable retention, MIME allow-list,
  optional virus-scan hook.
