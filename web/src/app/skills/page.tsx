"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { Sparkles, FolderOpen, Trash2, Download, MoveRight, Check } from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  getSkills,
  getSkill,
  deleteSkill,
  moveSkill,
  getStatus,
  type SkillInfo,
  type AgentInfo,
  type SkillDetail,
} from "@/lib/api";

export default function SkillsPage() {
  const [skills, setSkills] = useState<SkillInfo[]>([]);
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [moveError, setMoveError] = useState<string | null>(null);
  // Detail dialog: click any skill card to view its full SKILL.md.
  // We track three states:
  //   - selected:    summary card the user clicked (shown in dialog header)
  //   - detail:      full content fetched on open (null while loading)
  //   - detailError: surfaced if the GET fails
  const [selected, setSelected] = useState<SkillInfo | null>(null);
  const [detail, setDetail] = useState<SkillDetail | null>(null);
  const [detailError, setDetailError] = useState<string | null>(null);

  const fetchSkills = () => {
    setLoading(true);
    getSkills()
      .then(setSkills)
      .catch(() => setSkills([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchSkills();
    // Agents list drives the "Move to agent…" dropdown options.
    getStatus()
      .then((s) => setAgents(s.agents || []))
      .catch(() => setAgents([]));
  }, []);

  const handleDelete = async () => {
    if (!deleteTarget) return;
    await deleteSkill(deleteTarget);
    setDeleteTarget(null);
    fetchSkills();
  };

  // Returns true when this skill is already at the given scope, used to
  // render a check mark + disable the option in the dropdown.
  const isAtScope = (skill: SkillInfo, scope: string): boolean => {
    if (scope === "user") return skill.type === "user";
    if (scope.startsWith("agent:")) {
      return skill.type === "agent" && skill.owner === scope.slice("agent:".length);
    }
    return false;
  };

  const handleMove = async (skill: SkillInfo, scope: string) => {
    setMoveError(null);
    const res = await moveSkill(skill.name, scope);
    if (!res.ok) {
      setMoveError(res.error || "Move failed");
      return;
    }
    fetchSkills();
  };

  // Open the detail dialog and fetch the full SKILL.md content.
  // The fetch is fire-and-forget — the dialog opens immediately with
  // a loading state, then renders content (or an error) when the
  // network resolves.
  const openDetail = async (skill: SkillInfo) => {
    setSelected(skill);
    setDetail(null);
    setDetailError(null);
    try {
      const d = await getSkill(skill.name);
      setDetail(d);
    } catch (e) {
      setDetailError((e as Error)?.message || "Failed to load skill");
    }
  };

  const closeDetail = () => {
    setSelected(null);
    setDetail(null);
    setDetailError(null);
  };

  // Strip leading YAML frontmatter from SKILL.md before rendering. We
  // already display the metadata (name / description / location) in
  // the dialog header, so duplicating the raw `--- ... ---` block
  // would just be visual noise.
  const stripFrontmatter = (md: string): string => {
    if (!md.trimStart().startsWith("---")) return md;
    const idx = md.indexOf("\n---", md.indexOf("---") + 3);
    if (idx < 0) return md;
    return md.slice(idx + 4).replace(/^\s*\n/, "");
  };

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Skills</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Installed skills that agents can use
          </p>
        </div>
        <Button variant="outline">
          <Download className="h-4 w-4 mr-2" />
          Install Skill
        </Button>
      </div>

      {loading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => (
            <Skeleton key={i} className="h-40" />
          ))}
        </div>
      ) : skills.length === 0 ? (
        <div className="rounded-lg border border-border bg-card">
          <div className="flex flex-col items-center justify-center py-16">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 mb-4">
              <Sparkles className="h-7 w-7 text-primary" />
            </div>
            <p className="text-sm text-muted-foreground mb-1">No skills installed</p>
            <p className="text-xs text-muted-foreground/60">
              Skills extend agent capabilities with specialized behaviors
            </p>
          </div>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {skills.map((skill) => (
            <div
              key={skill.name}
              role="button"
              tabIndex={0}
              onClick={() => openDetail(skill)}
              onKeyDown={(e) => {
                if (e.key === "Enter" || e.key === " ") {
                  e.preventDefault();
                  openDetail(skill);
                }
              }}
              className="group rounded-lg border border-border bg-card p-5 transition-colors hover:bg-muted/50 cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary/40"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2.5">
                  <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
                    <Sparkles className="h-4 w-4 text-primary" />
                  </div>
                  <div>
                    <p className="text-sm font-medium">{skill.name}</p>
                    <div className="flex items-center gap-1 mt-1 flex-wrap">
                      {skill.builtin ? (
                        <Badge className="text-[10px] bg-primary/10 text-primary hover:bg-primary/20">
                          builtin
                        </Badge>
                      ) : (
                        <Badge variant="outline" className="text-[10px]">
                          {skill.type || "skill"}
                        </Badge>
                      )}
                      {skill.owner && (
                        <Badge variant="secondary" className="text-[10px]">
                          {skill.owner}
                        </Badge>
                      )}
                      {/* Kind badge: surfaces orchestrator-style skills
                          (protocol / suite) so they're recognizable at
                          a glance. Atomic skills (kind="" or "skill")
                          don't get an extra badge. */}
                      {skill.kind === "suite" && (
                        <Badge className="text-[10px] bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30 hover:bg-amber-500/20">
                          编排器
                        </Badge>
                      )}
                      {skill.kind === "protocol" && (
                        <Badge className="text-[10px] bg-violet-500/15 text-violet-700 dark:text-violet-300 border-violet-500/30 hover:bg-violet-500/20">
                          协议
                        </Badge>
                      )}
                    </div>
                  </div>
                </div>
                {!skill.builtin && (
                  // stopPropagation on every action button so the
                  // outer card's onClick (open detail) doesn't fire
                  // when the user is using the move/delete affordance.
                  <div
                    className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity"
                    onClick={(e) => e.stopPropagation()}
                  >
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        className="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
                        title="Move skill"
                      >
                        <MoveRight className="h-3.5 w-3.5" />
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="w-52">
                        <div className="px-1.5 py-1 text-[10px] uppercase tracking-wider text-muted-foreground/70 select-none">
                          Move to
                        </div>
                        <DropdownMenuItem
                          onClick={() => handleMove(skill, "user")}
                          disabled={isAtScope(skill, "user")}
                          className="flex items-center justify-between"
                        >
                          <div>
                            <div className="text-sm">User (shared)</div>
                            <div className="text-[10px] text-muted-foreground">
                              All agents can use it
                            </div>
                          </div>
                          {isAtScope(skill, "user") && <Check className="h-3.5 w-3.5 text-primary" />}
                        </DropdownMenuItem>
                        {agents.length > 0 && <DropdownMenuSeparator />}
                        {agents.map((a) => {
                          const scope = `agent:${a.id}`;
                          const at = isAtScope(skill, scope);
                          return (
                            <DropdownMenuItem
                              key={a.id}
                              onClick={() => handleMove(skill, scope)}
                              disabled={at}
                              className="flex items-center justify-between"
                            >
                              <div>
                                <div className="text-sm">Agent: {a.id}</div>
                                <div className="text-[10px] text-muted-foreground">
                                  Only this agent
                                </div>
                              </div>
                              {at && <Check className="h-3.5 w-3.5 text-primary" />}
                            </DropdownMenuItem>
                          );
                        })}
                      </DropdownMenuContent>
                    </DropdownMenu>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-7 w-7 text-muted-foreground hover:text-destructive"
                      onClick={() => setDeleteTarget(skill.name)}
                      title="Delete skill"
                    >
                      <Trash2 className="h-3.5 w-3.5" />
                    </Button>
                  </div>
                )}
              </div>
              <p className="text-sm text-muted-foreground line-clamp-2 mb-3">
                {skill.description || "No description"}
              </p>
              <div className="flex items-center gap-1.5 text-muted-foreground/60">
                <FolderOpen className="h-3 w-3" />
                <span className="text-[11px] font-mono truncate">
                  {skill.location}
                </span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Skill detail dialog — opens when a card is clicked. */}
      <Dialog open={!!selected} onOpenChange={(open) => { if (!open) closeDetail(); }}>
        <DialogContent className="sm:max-w-3xl max-h-[85vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-primary" />
              <span>{selected?.name}</span>
              {selected?.builtin && (
                <Badge className="text-[10px] bg-primary/10 text-primary hover:bg-primary/20">builtin</Badge>
              )}
              {selected && !selected.builtin && (
                <Badge variant="outline" className="text-[10px]">{selected.type || "skill"}</Badge>
              )}
              {selected?.owner && (
                <Badge variant="secondary" className="text-[10px]">{selected.owner}</Badge>
              )}
              {selected?.kind === "suite" && (
                <Badge className="text-[10px] bg-amber-500/15 text-amber-700 dark:text-amber-300 border-amber-500/30 hover:bg-amber-500/20">编排器</Badge>
              )}
              {selected?.kind === "protocol" && (
                <Badge className="text-[10px] bg-violet-500/15 text-violet-700 dark:text-violet-300 border-violet-500/30 hover:bg-violet-500/20">协议</Badge>
              )}
            </DialogTitle>
            {selected?.description && (
              <DialogDescription className="text-sm text-muted-foreground">
                {selected.description}
              </DialogDescription>
            )}
            {selected?.location && (
              <p className="text-[11px] font-mono text-muted-foreground/60 truncate">
                {selected.location}
              </p>
            )}
          </DialogHeader>

          <div className="flex-1 overflow-y-auto -mx-6 px-6 mt-2">
            {detailError ? (
              <p className="text-sm text-destructive">{detailError}</p>
            ) : !detail ? (
              <div className="space-y-2 py-4">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-5/6" />
                <Skeleton className="h-4 w-4/6" />
              </div>
            ) : (
              <article className="prose prose-sm dark:prose-invert max-w-none prose-headings:scroll-mt-4 prose-pre:text-[12px]">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {stripFrontmatter(detail.content)}
                </ReactMarkdown>
              </article>
            )}
          </div>
        </DialogContent>
      </Dialog>

      <AlertDialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Skill</AlertDialogTitle>
            <AlertDialogDescription>
              Remove <strong>{deleteTarget}</strong> from installed skills?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Remove
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Surface move errors (e.g. destination already has a skill with this name) */}
      <AlertDialog open={!!moveError} onOpenChange={() => setMoveError(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Could not move skill</AlertDialogTitle>
            <AlertDialogDescription>{moveError}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogAction onClick={() => setMoveError(null)}>OK</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
