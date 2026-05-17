"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Badge } from "@/components/ui/badge";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Cable, Plus, Trash2 } from "lucide-react";
import {
  getMcpServers,
  createMcpServer,
  deleteMcpServer,
  updateMcpServer,
  type McpServerInfo,
} from "@/lib/api";

export default function McpPage() {
  const [servers, setServers] = useState<McpServerInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [showAdd, setShowAdd] = useState(false);

  // Pre-filled Reasonix suggestion
  const [newServer, setNewServer] = useState<{
    name: string;
    type: "stdio" | "http";
    command: string;
    args: string;
    url: string;
  }>({
    name: "",
    type: "stdio",
    command: "",
    args: "",
    url: "",
  });

  const fetchServers = () => {
    setLoading(true);
    getMcpServers()
      .then(setServers)
      .catch(() => setServers([]))
      .finally(() => setLoading(false));
  };

  useEffect(() => {
    fetchServers();
  }, []);

  const handleCreate = async () => {
    const body: Record<string, unknown> = {
      name: newServer.name,
      type: newServer.type,
    };
    if (newServer.type === "stdio") {
      body.command = newServer.command;
      body.args = newServer.args
        .split(/\s+/)
        .filter((s) => s.length > 0);
    } else {
      body.url = newServer.url;
    }

    const res = await createMcpServer(body);
    if (res.ok) {
      setShowAdd(false);
      setNewServer({ name: "", type: "stdio", command: "", args: "", url: "" });
      fetchServers();
    } else {
      alert(res.error || "Failed to create server");
    }
  };

  const handleDelete = async (name: string) => {
    if (!confirm(`Delete MCP server "${name}"?`)) return;
    await deleteMcpServer(name);
    fetchServers();
  };

  const typeColor = (t: string) =>
    t === "stdio"
      ? "bg-violet-500/10 text-violet-600 dark:text-violet-400 border-violet-500/20"
      : "bg-cyan-500/10 text-cyan-600 dark:text-cyan-400 border-cyan-500/20";

  const statusDot = (s: string) => {
    if (s === "connected") return "bg-emerald-500";
    if (s === "disconnected") return "bg-red-400";
    return "bg-muted-foreground/30";
  };

  return (
    <div className="p-6 space-y-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-semibold tracking-tight">
            MCP Servers
          </h2>
          <p className="text-sm text-muted-foreground mt-1">
            Model Context Protocol servers extend agents with external tools.
          </p>
        </div>
        <Button onClick={() => setShowAdd(true)}>
          <Plus className="h-4 w-4 mr-2" />
          Add Server
        </Button>
        <Dialog open={showAdd} onOpenChange={setShowAdd}>
          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>Add MCP Server</DialogTitle>
              <DialogDescription>
                Connect an MCP server to make its tools available to all agents.
              </DialogDescription>
            </DialogHeader>

            <div className="space-y-4 py-2">
              <div className="space-y-2">
                <Label htmlFor="mcp-name">Name</Label>
                <Input
                  id="mcp-name"
                  placeholder="reasonix"
                  value={newServer.name}
                  onChange={(e) =>
                    setNewServer({ ...newServer, name: e.target.value })
                  }
                />
              </div>

              <div className="space-y-2">
                <Label>Transport</Label>
                <Select
                  value={newServer.type}
                  onValueChange={(v) =>
                    setNewServer({ ...newServer, type: v as "stdio" | "http" })
                  }
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="stdio">stdio (subprocess)</SelectItem>
                    <SelectItem value="http">http (remote)</SelectItem>
                  </SelectContent>
                </Select>
              </div>

              {newServer.type === "stdio" ? (
                <>
                  <div className="space-y-2">
                    <Label htmlFor="mcp-cmd">Command</Label>
                    <Input
                      id="mcp-cmd"
                      placeholder="pawnix"
                      value={newServer.command}
                      onChange={(e) =>
                        setNewServer({ ...newServer, command: e.target.value })
                      }
                    />
                  </div>
                  <div className="space-y-2">
                    <Label htmlFor="mcp-args">Arguments (space-separated)</Label>
                    <Input
                      id="mcp-args"
                      placeholder="mcp-reasonix"
                      value={newServer.args}
                      onChange={(e) =>
                        setNewServer({ ...newServer, args: e.target.value })
                      }
                    />
                  </div>
                </>
              ) : (
                <div className="space-y-2">
                  <Label htmlFor="mcp-url">URL</Label>
                  <Input
                    id="mcp-url"
                    placeholder="http://localhost:3001/mcp"
                    value={newServer.url}
                    onChange={(e) =>
                      setNewServer({ ...newServer, url: e.target.value })
                    }
                  />
                </div>
              )}
            </div>

            <DialogFooter>
              <Button variant="outline" onClick={() => setShowAdd(false)}>
                Cancel
              </Button>
              <Button onClick={handleCreate} disabled={!newServer.name}>
                Add Server
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {loading ? (
        <div className="space-y-3">
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
          <Skeleton className="h-10 w-full" />
        </div>
      ) : servers.length === 0 ? (
        <div className="rounded-lg border border-dashed p-12 text-center">
          <Cable className="h-8 w-8 mx-auto text-muted-foreground/40 mb-3" />
          <p className="text-sm text-muted-foreground mb-4">
            No MCP servers configured yet
          </p>
          <Button variant="outline" onClick={() => setShowAdd(true)}>
            <Plus className="h-4 w-4 mr-2" />
            Add your first MCP server
          </Button>
        </div>
      ) : (
        <div className="rounded-lg border">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Status</TableHead>
                <TableHead>Name</TableHead>
                <TableHead>Transport</TableHead>
                <TableHead>Details</TableHead>
                <TableHead className="w-20">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {servers.map((srv) => (
                <TableRow key={srv.name}>
                  <TableCell>
                    <span
                      className={`inline-block h-2.5 w-2.5 rounded-full ${statusDot(srv.status)}`}
                      title={srv.status}
                    />
                  </TableCell>
                  <TableCell>
                    <span className="font-medium">{srv.name}</span>
                    {srv.toolCount !== undefined && srv.toolCount >= 0 && (
                      <span className="ml-2 text-xs text-muted-foreground">
                        ({srv.toolCount} tools)
                      </span>
                    )}
                  </TableCell>
                  <TableCell>
                    <Badge variant="outline" className={typeColor(srv.type)}>
                      {srv.type}
                    </Badge>
                  </TableCell>
                  <TableCell className="text-sm text-muted-foreground font-mono">
                    {srv.type === "stdio"
                      ? `${srv.command || "—"} ${(srv.args || []).join(" ")}`
                      : srv.url || "—"}
  
                  </TableCell>
                  <TableCell>
                    <Button
                      variant="ghost"
                      size="icon"
                      onClick={() => handleDelete(srv.name)}
                      title="Delete server"
                    >
                      <Trash2 className="h-4 w-4 text-muted-foreground hover:text-destructive" />
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>
      )}
    </div>
  );
}
