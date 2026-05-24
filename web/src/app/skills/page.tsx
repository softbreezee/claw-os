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
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import {
  Sparkles, Search, X,
  FolderOpen, Trash2, Download,
} from "lucide-react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import {
  getSkills,
  getSkill,
  deleteSkill,
  getStatus,
  updateSkillScope,
  type SkillInfo,
  type AgentInfo,
  type SkillDetail,
} from "@/lib/api";

export default function SkillsPage() {
  const [skills, setSkills] = useState<SkillInfo[]>([]);
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState("");
  const [activeTags, setActiveTags] = useState<string[]>([]);
  const [selected, setSelected] = useState<SkillInfo | null>(null);
  const [detail, setDetail] = useState<SkillDetail | null>(null);
  const [detailError, setDetailError] = useState<string | null>(null);

  const builtinSkills = skills.filter((s) => s.builtin || s.type === "builtin");
  const customSkills = skills.filter((s) => !s.builtin && s.type !== "builtin");

  // All unique tags across custom skills
  const allTags = [...new Set(customSkills.flatMap((s) => s.tags || []))].sort();

  // Filtered custom skills by search + tags
  const filteredCustom = customSkills.filter((s) => {
    const queryMatch = !searchQuery || s.name.toLowerCase().includes(searchQuery.toLowerCase()) || (s.description || "").toLowerCase().includes(searchQuery.toLowerCase());
    const tagMatch = activeTags.length === 0 || (activeTags.every((t) => (s.tags || []).includes(t)));
    return queryMatch && tagMatch;
  });

  const fetchSkills = () => {
    setLoading(true);
    getSkills()
      .then(setSkills)
      .catch(() => setSkills([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchSkills();
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

  const handleScopeChange = async (skill: SkillInfo, agentId: string, add: boolean) => {
    const agents = skill.agents || [];
    let newAgents: string[];
    if (add) {
      newAgents = [...agents, agentId];
    } else {
      newAgents = agents.filter((a) => a !== agentId);
    }
    await updateSkillScope(skill.name, newAgents);
    fetchSkills();
  };

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

  const stripFrontmatter = (md: string): string => {
    if (!md.trimStart().startsWith("---")) return md;
    const idx = md.indexOf("\n---", md.indexOf("---") + 3);
    if (idx < 0) return md;
    return md.slice(idx + 4).replace(/^\s*\n/, "");
  };

  const isCommon = (skill: SkillInfo) =>
    !skill.agents || skill.agents.length === 0;

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Skills</h2>
          <p className="text-sm text-muted-foreground mt-1">
            {builtinSkills.length} built-in · {customSkills.length} custom
          </p>
        </div>
        <Button variant="outline">
          <Download className="h-4 w-4 mr-2" />
          Install Skill
        </Button>
      </div>

      {/* Search + tag filter */}
      {!loading && skills.length > 0 && (
        <div className="flex flex-col gap-3">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground/60" />
            <input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder="Search skills..."
              className="w-full h-10 rounded-lg border border-border bg-card pl-9 pr-4 text-sm outline-none focus:ring-2 focus:ring-primary/30 transition-colors"
            />
            {searchQuery && (
              <button onClick={() => setSearchQuery("")} className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground/40 hover:text-foreground">
                <X className="h-4 w-4" />
              </button>
            )}
          </div>
          <div className="flex flex-wrap gap-1.5">
            {allTags.map((tag) => {
              const active = activeTags.includes(tag);
              return (
                <button
                  key={tag}
                  onClick={() => setActiveTags(active ? activeTags.filter((t) => t !== tag) : [...activeTags, tag])}
                  className={`inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium border transition-colors ${
                    active
                      ? "bg-primary/10 text-primary border-primary/30"
                      : "bg-muted/30 text-muted-foreground border-border hover:bg-muted/60"
                  }`}
                >
                  {tag}
                  {active && <X className="h-3 w-3" />}
                </button>
              );
            })}
          </div>
        </div>
      )}

      {loading ? (
        <div className="space-y-3">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      ) : skills.length === 0 ? (
        <div className="rounded-lg border border-border bg-card">
          <div className="flex flex-col items-center justify-center py-16">
            <Sparkles className="h-8 w-8 text-muted-foreground/40 mb-3" />
            <p className="text-sm text-muted-foreground">No skills found</p>
          </div>
        </div>
      ) : (
        <>
          {/* Built-in */}
          <div className="rounded-lg border">
            <div className="px-4 py-2.5 border-b bg-muted/20">
              <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground/60">
                Built-in ({builtinSkills.length}) — always loaded for all agents
              </p>
            </div>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Description</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {builtinSkills.length === 0 ? (
                  <TableRow><TableCell colSpan={2} className="text-sm text-muted-foreground italic">None</TableCell></TableRow>
                ) : builtinSkills.map((skill) => (
                  <TableRow key={skill.name} className="cursor-pointer hover:bg-muted/50" onClick={() => openDetail(skill)}>
                    <TableCell>
                      <div className="flex items-center gap-1.5">
                        <span className="font-medium text-sm">{skill.name}</span>
                        <Badge className="text-[10px] bg-primary/10 text-primary">builtin</Badge>
                        {(skill.tags || []).map((t) => (
                          <span key={t} className="text-[10px] text-muted-foreground/50 font-mono">#{t}</span>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground max-w-lg truncate">
                      {skill.description || "—"}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Custom */}
          <div className="rounded-lg border">
            <div className="px-4 py-2.5 border-b bg-muted/20">
              <p className="text-xs font-semibold uppercase tracking-wider text-muted-foreground/60">
                Custom ({customSkills.length}) — assign to agents
              </p>
            </div>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead className="w-64">Assigned To</TableHead>
                  <TableHead className="w-20">Actions</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredCustom.length === 0 ? (
                  <TableRow><TableCell colSpan={4} className="text-sm text-muted-foreground italic">No skills match your search</TableCell></TableRow>
                ) : filteredCustom.map((skill) => (
                  <TableRow key={skill.name}>
                    <TableCell>
                      <div className="flex items-center gap-1.5">
                        <span className="font-medium text-sm cursor-pointer hover:text-primary" onClick={() => openDetail(skill)}>
                          {skill.name}
                        </span>
                        {(skill.tags || []).map((t) => (
                          <span key={t} className="text-[10px] text-muted-foreground/50 font-mono">#{t}</span>
                        ))}
                      </div>
                    </TableCell>
                    <TableCell className="text-sm text-muted-foreground max-w-xs truncate">
                      {skill.description || "—"}
                    </TableCell>
                    <TableCell>
                      <div className="flex flex-wrap items-center gap-1.5">
                        {isCommon(skill) ? (
                          <Badge variant="secondary" className="text-[10px] bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border-emerald-500/20">
                            Common
                          </Badge>
                        ) : (
                          skill.agents?.map((a) => (
                            <Badge
                              key={a}
                              variant="secondary"
                              className="text-[10px] pr-1 gap-1"
                            >
                              {a}
                              <button
                                onClick={(e) => { e.stopPropagation(); handleScopeChange(skill, a, false); }}
                                className="hover:text-destructive transition-colors"
                              >
                                ✕
                              </button>
                            </Badge>
                          ))
                        )}
                        <Select
                          value=""
                          onValueChange={(v) => {
                            if (v === "__common__") {
                              // Make common: clear agents list
                              updateSkillScope(skill.name, []).then(() => fetchSkills());
                            } else {
                              if (v) handleScopeChange(skill, v, true);
                            }
                          }}
                        >
                          <SelectTrigger className="h-6 w-6 border-dashed border-muted-foreground/30 [&>svg]:hidden">
                            <span className="text-xs text-muted-foreground/60">+</span>
                          </SelectTrigger>
                          <SelectContent>
                            <SelectItem value="__common__">Make Common</SelectItem>
                            {agents
                              .filter((a) => !skill.agents?.includes(a.id))
                              .map((a) => (
                                <SelectItem key={a.id} value={a.id}>
                                  Assign: {a.id}
                                </SelectItem>
                              ))}
                          </SelectContent>
                        </Select>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-7 w-7 text-muted-foreground hover:text-destructive"
                        onClick={() => setDeleteTarget(skill.name)}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>
        </>
      )}

      {/* Detail dialog */}
      <Dialog open={!!selected} onOpenChange={(open) => { if (!open) closeDetail(); }}>
        <DialogContent className="sm:max-w-3xl max-h-[85vh] overflow-hidden flex flex-col">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Sparkles className="h-4 w-4 text-primary" />
              {selected?.name}
              {selected?.builtin && <Badge className="text-[10px] bg-primary/10 text-primary">builtin</Badge>}
            </DialogTitle>
            {selected?.description && (
              <DialogDescription>{selected.description}</DialogDescription>
            )}
          </DialogHeader>

          {/* Tags editor — only for non-builtin */}
          {selected && !selected.builtin && (
            <div className="px-6 py-2 shrink-0 border-b">
              <div className="flex flex-wrap items-center gap-1.5">
                {(selected.tags || []).map((t) => (
                  <span key={t} className="inline-flex items-center gap-1 rounded-full bg-primary/10 text-primary text-xs px-2 py-0.5">
                    #{t}
                    <button
                      onClick={() => {
                        const newTags = (selected.tags || []).filter((x) => x !== t);
                        updateSkillScope(selected.name, selected.agents || [], newTags).then(() => {
                          setSkills((prev) => prev.map((s) => s.name === selected.name ? { ...s, tags: newTags } : s));
                          setSelected({ ...selected, tags: newTags });
                        });
                      }}
                      className="hover:text-destructive"
                    >✕</button>
                  </span>
                ))}
                {/* Add tag input */}
                <input
                  placeholder="+ tag"
                  className="inline-flex h-6 w-20 text-xs border-0 border-b border-dashed border-muted-foreground/30 bg-transparent outline-none focus:border-primary focus:w-28 transition-all"
                  onKeyDown={(e) => {
                    if (e.key === "Enter" && e.currentTarget.value.trim()) {
                      const tag = e.currentTarget.value.trim();
                      const newTags = [...(selected.tags || []), tag];
                      updateSkillScope(selected.name, selected.agents || [], newTags).then(() => {
                        setSkills((prev) => prev.map((s) => s.name === selected.name ? { ...s, tags: newTags } : s));
                        setSelected({ ...selected, tags: newTags });
                      });
                      e.currentTarget.value = "";
                    }
                  }}
                />
              </div>
            </div>
          )}

          <div className="flex-1 overflow-y-auto -mx-6 px-6">
            {detailError ? (
              <p className="text-sm text-destructive">{detailError}</p>
            ) : !detail ? (
              <div className="space-y-2 py-4">
                <Skeleton className="h-4 w-full" />
                <Skeleton className="h-4 w-5/6" />
                <Skeleton className="h-4 w-4/6" />
              </div>
            ) : (
              <article className="prose prose-sm dark:prose-invert max-w-none prose-headings:scroll-mt-4">
                <ReactMarkdown remarkPlugins={[remarkGfm]}>
                  {stripFrontmatter(detail.content)}
                </ReactMarkdown>
              </article>
            )}
          </div>
        </DialogContent>
      </Dialog>

      {/* Delete */}
      <AlertDialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove Skill</AlertDialogTitle>
            <AlertDialogDescription>
              Remove <strong>{deleteTarget}</strong>?
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction onClick={handleDelete} className="bg-destructive">Remove</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
