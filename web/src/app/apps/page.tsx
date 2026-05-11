"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Skeleton } from "@/components/ui/skeleton";
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
import { AppWindow, ExternalLink, Plus, Pencil, Trash2 } from "lucide-react";
import { getApps, createApp, updateApp, deleteApp, type AppEntry } from "@/lib/api";

// Apps page — a tiny bookmark manager for external web tools the user
// wants reachable from inside fastclaw with one click. We deliberately
// open the URL in a new tab rather than embed via iframe: surge.sh
// pages and most modern web apps set X-Frame-Options / CSP that block
// embedding, and a hard-disabled iframe is a worse UX than just opening
// a fresh tab.

export default function AppsPage() {
  const [apps, setApps] = useState<AppEntry[]>([]);
  const [loading, setLoading] = useState(true);

  // Edit dialog state. editingName==null & dialogOpen=true → "add" mode.
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingName, setEditingName] = useState<string | null>(null);
  const [formName, setFormName] = useState("");
  const [formURL, setFormURL] = useState("");
  const [formDesc, setFormDesc] = useState("");
  const [formError, setFormError] = useState<string | null>(null);

  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);

  const fetchApps = () => {
    setLoading(true);
    getApps()
      .then((list) => setApps(list || []))
      .catch(() => setApps([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchApps();
  }, []);

  const openAddDialog = () => {
    setEditingName(null);
    setFormName("");
    setFormURL("");
    setFormDesc("");
    setFormError(null);
    setDialogOpen(true);
  };

  const openEditDialog = (app: AppEntry) => {
    setEditingName(app.name);
    setFormName(app.name);
    setFormURL(app.url);
    setFormDesc(app.description || "");
    setFormError(null);
    setDialogOpen(true);
  };

  const handleSave = async () => {
    setFormError(null);
    const payload: AppEntry = {
      name: formName.trim(),
      url: formURL.trim(),
      description: formDesc.trim() || undefined,
    };
    if (!payload.name || !payload.url) {
      setFormError("Name and URL are required");
      return;
    }
    if (!/^https?:\/\//i.test(payload.url)) {
      setFormError("URL must start with http:// or https://");
      return;
    }
    const res = editingName
      ? await updateApp(editingName, payload)
      : await createApp(payload);
    if (!res?.ok) {
      setFormError(res?.error || "Save failed");
      return;
    }
    setDialogOpen(false);
    fetchApps();
  };

  const handleDelete = async () => {
    if (!deleteTarget) return;
    await deleteApp(deleteTarget);
    setDeleteTarget(null);
    fetchApps();
  };

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">Apps</h2>
          <p className="text-sm text-muted-foreground mt-1">
            Quick links to external web tools. Opens in a new tab.
          </p>
        </div>
        <Button variant="outline" onClick={openAddDialog}>
          <Plus className="h-4 w-4 mr-2" />
          Add App
        </Button>
      </div>

      {loading ? (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {[1, 2, 3].map((i) => <Skeleton key={i} className="h-32" />)}
        </div>
      ) : apps.length === 0 ? (
        <div className="rounded-lg border border-border bg-card">
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary/10 mb-4">
              <AppWindow className="h-7 w-7 text-primary" />
            </div>
            <p className="text-sm text-muted-foreground mb-1">No apps yet</p>
            <p className="text-xs text-muted-foreground/60 mb-4">
              Add a quick link to any external dashboard or tool
            </p>
            <Button variant="outline" size="sm" onClick={openAddDialog}>
              <Plus className="h-4 w-4 mr-2" />
              Add App
            </Button>
          </div>
        </div>
      ) : (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {apps.map((app) => (
            <a
              key={app.name}
              href={app.url}
              target="_blank"
              rel="noopener noreferrer"
              className="group relative block rounded-lg border border-border bg-card p-5 transition-colors hover:bg-muted/50 hover:border-primary/40"
            >
              <div className="flex items-start justify-between mb-3">
                <div className="flex items-center gap-2.5 min-w-0">
                  <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary/10">
                    <AppWindow className="h-4 w-4 text-primary" />
                  </div>
                  <div className="min-w-0">
                    <p className="text-sm font-medium truncate">{app.name}</p>
                    <p className="text-[11px] text-muted-foreground/60 font-mono truncate">
                      {app.url.replace(/^https?:\/\//, "")}
                    </p>
                  </div>
                </div>
                <ExternalLink className="h-3.5 w-3.5 text-muted-foreground/60 shrink-0" />
              </div>

              {app.description && (
                <p className="text-xs text-muted-foreground line-clamp-3 mb-2">
                  {app.description}
                </p>
              )}

              {/* Hover actions — overlay top-right, stop link from firing. */}
              <div
                className="absolute top-2 right-2 flex items-center gap-0.5 opacity-0 group-hover:opacity-100 transition-opacity"
                onClick={(e) => { e.preventDefault(); e.stopPropagation(); }}
              >
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-muted-foreground hover:text-foreground"
                  onClick={() => openEditDialog(app)}
                  title="Edit"
                >
                  <Pencil className="h-3.5 w-3.5" />
                </Button>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-7 w-7 text-muted-foreground hover:text-destructive"
                  onClick={() => setDeleteTarget(app.name)}
                  title="Remove"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                </Button>
              </div>
            </a>
          ))}
        </div>
      )}

      {/* Add / Edit dialog */}
      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>{editingName ? "Edit App" : "Add App"}</DialogTitle>
            <DialogDescription>
              External web tool reachable from the sidebar. Opens in a new tab.
            </DialogDescription>
          </DialogHeader>
          <div className="space-y-4 py-2">
            <div className="space-y-2">
              <Label>Name</Label>
              <Input
                value={formName}
                onChange={(e) => setFormName(e.target.value)}
                placeholder="Signal Agent V8.2"
              />
            </div>
            <div className="space-y-2">
              <Label>URL</Label>
              <Input
                value={formURL}
                onChange={(e) => setFormURL(e.target.value)}
                placeholder="https://example.surge.sh/"
                className="font-mono text-sm"
              />
            </div>
            <div className="space-y-2">
              <Label>Description (optional)</Label>
              <Input
                value={formDesc}
                onChange={(e) => setFormDesc(e.target.value)}
                placeholder="What this tool does"
              />
            </div>
            {formError && (
              <p className="text-sm text-destructive">{formError}</p>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDialogOpen(false)}>
              Cancel
            </Button>
            <Button onClick={handleSave}>
              {editingName ? "Update" : "Add"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete confirmation */}
      <AlertDialog open={!!deleteTarget} onOpenChange={() => setDeleteTarget(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove App</AlertDialogTitle>
            <AlertDialogDescription>
              Remove <strong>{deleteTarget}</strong> from your apps?
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
    </div>
  );
}
