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
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Skeleton } from "@/components/ui/skeleton";
import { Sparkles, FolderOpen, Trash2, Download, MoveRight, Check } from "lucide-react";
import {
  getSkills,
  deleteSkill,
  moveSkill,
  getStatus,
  type SkillInfo,
  type AgentInfo,
} from "@/lib/api";

export default function SkillsPage() {
  const [skills, setSkills] = useState<SkillInfo[]>([]);
  const [agents, setAgents] = useState<AgentInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [moveError, setMoveError] = useState<string | null>(null);

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
              className="group rounded-lg border border-border bg-card p-5 transition-colors hover:bg-muted/50"
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
                    </div>
                  </div>
                </div>
                {!skill.builtin && (
                  <div className="flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity">
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        className="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground transition-colors"
                        title="Move skill"
                      >
                        <MoveRight className="h-3.5 w-3.5" />
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="end" className="w-52">
                        <DropdownMenuLabel className="text-[10px] uppercase tracking-wider text-muted-foreground/70">
                          Move to
                        </DropdownMenuLabel>
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
