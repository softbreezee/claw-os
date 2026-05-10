"use client";

import { useEffect, useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Badge } from "@/components/ui/badge";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
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
import { ChevronDown } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Skeleton } from "@/components/ui/skeleton";
import { Bot, Plus, Pencil, Trash2, FolderOpen, Check } from "lucide-react";
import { getAgents, getConfig, createAgent, updateAgent, deleteAgent, restartDaemon, waitForGateway, type AgentDetail } from "@/lib/api";

interface ModelGroup {
  group: string;
  models: string[];
}

/**
 * Combobox: input box + dropdown of models pulled live from the user's
 * configured providers (Models page).
 *
 * Why a combobox and not a plain Select?
 *   - Users sometimes want to pin an agent to a model that isn't yet in
 *     their provider config (e.g. they're about to add it). The free-form
 *     input lets them type any value, which is sent to the LLM API as-is.
 *
 * The previous implementation used a hard-coded SUGGESTED_MODELS list,
 * which meant the dropdown was disconnected from what the user actually
 * configured on the Models page. Fixed by accepting `groups` as a prop
 * sourced from getConfig().providers.
 */
function ModelCombobox({
  value,
  onChange,
  groups,
}: {
  value: string;
  onChange: (v: string) => void;
  groups: ModelGroup[];
}) {
  const [open, setOpen] = useState(false);
  const [inputVal, setInputVal] = useState(value);
  // Track whether the user is actively typing — only then do we filter
  // the dropdown by inputVal. Otherwise selecting a model and re-opening
  // the dropdown would only show that single model (the original bug).
  const [isTyping, setIsTyping] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => { setInputVal(value); }, [value]);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
        setIsTyping(false);
      }
    };
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  const handleInput = (v: string) => {
    setInputVal(v);
    onChange(v);
    setIsTyping(true);
    setOpen(true);
  };

  const handleSelect = (m: string) => {
    setInputVal(m);
    onChange(m);
    setIsTyping(false);
    setOpen(false);
  };

  const flatModels = groups.flatMap((g) => g.models);
  const filtered = isTyping && inputVal
    ? groups
        .map((g) => ({
          ...g,
          models: g.models.filter((m) => m.toLowerCase().includes(inputVal.toLowerCase())),
        }))
        .filter((g) => g.models.length > 0)
    : groups;

  return (
    <div ref={ref} className="relative">
      <div className="relative">
        <Input
          value={inputVal}
          onChange={(e) => handleInput(e.target.value)}
          onFocus={() => { setOpen(true); setIsTyping(false); }}
          placeholder="e.g. deepseek-v4-flash"
          className="pr-8 font-mono text-sm"
        />
        <button
          type="button"
          onClick={() => { setOpen(!open); setIsTyping(false); }}
          className="absolute right-2 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground transition-colors"
        >
          <ChevronDown className={`h-4 w-4 transition-transform ${open ? "rotate-180" : ""}`} />
        </button>
      </div>
      {open && (
        <div className="absolute z-50 mt-1 w-full rounded-lg border border-border bg-popover shadow-lg overflow-hidden">
          <div className="max-h-64 overflow-y-auto py-1">
            {groups.length === 0 && (
              <p className="px-3 py-3 text-xs text-muted-foreground italic">
                No models configured yet. Go to the <span className="font-semibold">Models</span> page to add a provider first.
              </p>
            )}
            {filtered.map((group) => (
              <div key={group.group}>
                <p className="px-3 pt-2 pb-1 text-[10px] font-semibold uppercase tracking-widest text-muted-foreground/60">
                  {group.group}
                </p>
                {group.models.map((m) => (
                  <button
                    key={m}
                    type="button"
                    onClick={() => handleSelect(m)}
                    className="flex w-full items-center gap-2 px-3 py-1.5 text-sm font-mono hover:bg-muted/60 transition-colors text-left"
                  >
                    <Check className={`h-3.5 w-3.5 shrink-0 ${value === m ? "text-primary opacity-100" : "opacity-0"}`} />
                    {m}
                  </button>
                ))}
              </div>
            ))}
            {groups.length > 0 && filtered.length === 0 && (
              <p className="px-3 py-2 text-xs text-muted-foreground italic">
                No matches — press Enter to use &quot;{inputVal}&quot; as a custom model
              </p>
            )}
          </div>
          {inputVal && !flatModels.includes(inputVal) && (
            <div className="border-t border-border px-3 py-2">
              <p className="text-[10px] text-muted-foreground/60">
                Custom model — will be sent as-is to the API
              </p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

export default function AgentsPage() {
  const [agents, setAgents] = useState<AgentDetail[]>([]);
  const [loading, setLoading] = useState(true);
  const [editAgent, setEditAgent] = useState<AgentDetail | null>(null);
  const [editOpen, setEditOpen] = useState(false);
  const [createOpen, setCreateOpen] = useState(false);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);
  const [restarting, setRestarting] = useState(false);

  // Live model groups from the user's configured providers (Models page).
  // Each provider becomes a group; entries are emitted as `${name}/${id}`
  // to mirror the Default Model selector convention.
  const [modelGroups, setModelGroups] = useState<ModelGroup[]>([]);

  const [newName, setNewName] = useState("");
  const [newModel, setNewModel] = useState("");
  const [newSoul, setNewSoul] = useState("");

  const [editModel, setEditModel] = useState("");
  const [editSoul, setEditSoul] = useState("");

  const fetchAgents = () => {
    setLoading(true);
    getAgents()
      .then(setAgents)
      .catch(() => setAgents([]))
      .finally(() => setLoading(false));
  };

  // Pull provider/model config and project it into combobox groups.
  // Re-run when dialogs open so newly added models show up without
  // requiring a full page refresh.
  const fetchModelOptions = () => {
    getConfig()
      .then((cfg) => {
        const groups: ModelGroup[] = [];
        const providers = cfg.providers || {};
        for (const [name, p] of Object.entries(providers)) {
          const models = (p.models || [])
            .filter((m) => m.id)
            .map((m) => `${name}/${m.id}`);
          if (models.length > 0) {
            groups.push({ group: name, models });
          }
        }
        setModelGroups(groups);

        // Seed newModel with the first available option so the create
        // dialog is usable out of the box.
        setNewModel((prev) => {
          if (prev) return prev;
          if (groups.length > 0 && groups[0].models.length > 0) {
            return groups[0].models[0];
          }
          return prev;
        });
      })
      .catch(() => setModelGroups([]));
  };

  useEffect(() => {
    fetchAgents();
    fetchModelOptions();
  }, []);

  // Refresh model options whenever a dialog opens — user may have
  // visited the Models page in another tab to add new models.
  useEffect(() => {
    if (createOpen || editOpen) fetchModelOptions();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [createOpen, editOpen]);

  const handleCreate = async () => {
    if (!newName.trim()) return;
    setSaving(true);
    await createAgent({ id: newName.trim(), model: newModel, soul: newSoul });
    setCreateOpen(false);
    setNewName("");
    // Reset to first available model after creation
    setNewModel(modelGroups[0]?.models[0] || "");
    setNewSoul("");
    setSaving(false);

    // Restart daemon so the new agent is loaded, then refresh the list
    setRestarting(true);
    await restartDaemon();
    await waitForGateway(15000);
    setRestarting(false);
    fetchAgents();
  };

  const handleEdit = (agent: AgentDetail) => {
    setEditAgent(agent);
    setEditModel(agent.model);
    setEditSoul(agent.soul || "");
    setEditOpen(true);
  };

  const handleSave = async () => {
    if (!editAgent) return;
    setSaving(true);
    await updateAgent(editAgent.id, { model: editModel, soul: editSoul });
    setEditOpen(false);
    setSaving(false);
    fetchAgents();
  };

  const handleDelete = async () => {
    if (!deleteId) return;
    await deleteAgent(deleteId);
    setDeleteId(null);
    fetchAgents();
  };

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      {restarting && (
        <div className="flex items-center gap-3 rounded-xl border border-primary/30 bg-primary/5 px-4 py-3 text-sm">
          <div className="h-4 w-4 shrink-0 rounded-full border-2 border-primary border-t-transparent animate-spin" />
          <span className="text-primary font-medium">Restarting gateway to load new agent…</span>
        </div>
      )}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Agents</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Manage your AI agents and their configurations
          </p>
        </div>
        <Button onClick={() => setCreateOpen(true)}>
          <Plus className="h-4 w-4 mr-2" />
          New Agent
        </Button>
      </div>

      <div className="rounded-lg border border-border bg-card">
        {loading ? (
          <div className="p-6 space-y-3">
            {[1, 2].map((i) => (
              <Skeleton key={i} className="h-14 w-full" />
            ))}
          </div>
        ) : agents.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 mb-4">
              <Bot className="h-7 w-7 text-primary" />
            </div>
            <p className="text-sm text-muted-foreground">No agents configured yet</p>
            <Button
              onClick={() => setCreateOpen(true)}
              variant="outline"
              className="mt-4"
            >
              Create your first agent
            </Button>
          </div>
        ) : (
          <div className="overflow-x-auto -mx-6 px-6">
          <Table>
            <TableHeader>
              <TableRow className="hover:bg-transparent">
                <TableHead>Name</TableHead>
                <TableHead>Model</TableHead>
                <TableHead>Workspace</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {agents.map((agent) => (
                <TableRow
                  key={agent.id}
                  className="cursor-pointer hover:bg-muted/50 transition-colors"
                  onClick={() => handleEdit(agent)}
                >
                  <TableCell>
                    <div className="flex items-center gap-3">
                      <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
                        <Bot className="h-4 w-4 text-primary" />
                      </div>
                      <span className="font-medium">
                        {agent.id}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <code className="bg-muted px-2 py-0.5 rounded font-mono text-xs">
                      {agent.model}
                    </code>
                  </TableCell>
                  <TableCell>
                    <div className="flex items-center gap-1.5 text-muted-foreground">
                      <FolderOpen className="h-3.5 w-3.5" />
                      <span className="text-xs font-mono truncate max-w-48">
                        {agent.workspace}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell>
                    <Badge
                      variant="outline"
                      className="bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 border-emerald-500/20"
                    >
                      Active
                    </Badge>
                  </TableCell>
                  <TableCell className="text-right">
                    <div className="flex items-center justify-end gap-1">
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-foreground"
                        onClick={(e) => {
                          e.stopPropagation();
                          handleEdit(agent);
                        }}
                      >
                        <Pencil className="h-3.5 w-3.5" />
                      </Button>
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        onClick={(e) => {
                          e.stopPropagation();
                          setDeleteId(agent.id);
                        }}
                      >
                        <Trash2 className="h-3.5 w-3.5" />
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
          </div>
        )}
      </div>

      {/* Create Dialog */}
      <Dialog open={createOpen} onOpenChange={setCreateOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Create New Agent</DialogTitle>
            <DialogDescription>
              Configure a new AI agent for your gateway
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Agent Name</Label>
              <Input
                value={newName}
                onChange={(e) => setNewName(e.target.value)}
                placeholder="my-agent"
              />
            </div>
            <div className="space-y-2">
              <Label>Model</Label>
              <ModelCombobox value={newModel} onChange={setNewModel} groups={modelGroups} />
            </div>
            <div className="space-y-2">
              <Label>Personality (SOUL.md)</Label>
              <Textarea
                value={newSoul}
                onChange={(e) => setNewSoul(e.target.value)}
                placeholder="You are a helpful AI assistant..."
                rows={4}
                className="resize-none"
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setCreateOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleCreate}
              disabled={!newName.trim() || saving}
            >
              {saving ? "Creating..." : "Create Agent"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Edit Dialog */}
      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent className="sm:max-w-2xl">
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <Bot className="h-5 w-5 text-primary" />
              {editAgent?.id}
            </DialogTitle>
            <DialogDescription>
              Edit agent configuration
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>Model</Label>
                <ModelCombobox value={editModel} onChange={setEditModel} groups={modelGroups} />
              </div>
              <div className="space-y-2">
                <Label>Workspace</Label>
                <Input
                  value={editAgent?.workspace || ""}
                  disabled
                  className="font-mono text-xs opacity-60"
                />
              </div>
            </div>
            <div className="space-y-2">
              <Label>Personality (SOUL.md)</Label>
              <Textarea
                value={editSoul}
                onChange={(e) => setEditSoul(e.target.value)}
                placeholder="Your personality and behavioral guidelines..."
                rows={8}
                className="resize-none font-mono text-sm min-h-[200px]"
              />
            </div>
          </div>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setEditOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={handleSave}
              disabled={saving}
            >
              {saving ? "Saving..." : "Save Changes"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation */}
      <AlertDialog open={!!deleteId} onOpenChange={() => setDeleteId(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Agent</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete <strong>{deleteId}</strong>?
              This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
