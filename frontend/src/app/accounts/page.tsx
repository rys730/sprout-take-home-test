"use client";

import { useEffect, useState } from "react";
import { SearchBar } from "@/components/search-bar";
import { Button } from "@/components/ui/button";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Plus, ChevronRight, ChevronDown, ChevronUp } from "lucide-react";
import { accountsApi } from "@/lib/api/accounts";
import type { AccountTreeNode } from "@/lib/types";

// Flatten children (excluding the root itself) into rows with depth
function flattenChildren(
  children: AccountTreeNode[],
  expandedIds: Set<string>,
  startDepth: number = 0
): { node: AccountTreeNode; depth: number; hasChildren: boolean }[] {
  const rows: { node: AccountTreeNode; depth: number; hasChildren: boolean }[] = [];

  function walk(list: AccountTreeNode[], depth: number) {
    for (const n of list) {
      const hasChildren = !!n.children && n.children.length > 0;
      rows.push({ node: n, depth, hasChildren });
      if (hasChildren && expandedIds.has(n.account.id)) {
        walk(n.children!, depth + 1);
      }
    }
  }

  walk(children, startDepth);
  return rows;
}

// Filter tree nodes by search query (code or name)
function filterTree(
  nodes: AccountTreeNode[],
  query: string
): AccountTreeNode[] {
  if (!query) return nodes;
  const q = query.toLowerCase();

  function matches(node: AccountTreeNode): boolean {
    return (
      node.account.code.toLowerCase().includes(q) ||
      node.account.name.toLowerCase().includes(q)
    );
  }

  function filter(list: AccountTreeNode[]): AccountTreeNode[] {
    const result: AccountTreeNode[] = [];
    for (const n of list) {
      const filteredChildren = n.children ? filter(n.children) : [];
      if (matches(n) || filteredChildren.length > 0) {
        result.push({
          ...n,
          children: filteredChildren.length > 0 ? filteredChildren : n.children,
        });
      }
    }
    return result;
  }

  return filter(nodes);
}

// Collect all node IDs in a tree (for expand all)
function collectIds(nodes: AccountTreeNode[]): string[] {
  const ids: string[] = [];
  function walk(list: AccountTreeNode[]) {
    for (const n of list) {
      ids.push(n.account.id);
      if (n.children) walk(n.children);
    }
  }
  walk(nodes);
  return ids;
}

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

// ---- Table for a single parent account ----
function AccountTable({
  parent,
  expandedIds,
  onToggle,
  isTableExpanded,
  onToggleTable,
}: {
  parent: AccountTreeNode;
  expandedIds: Set<string>;
  onToggle: (id: string) => void;
  isTableExpanded: boolean;
  onToggleTable: () => void;
}) {
  const children = parent.children ?? [];
  const rows = flattenChildren(children, expandedIds, 1);

  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow className="bg-muted/50 cursor-pointer" onClick={onToggleTable}>
              <TableHead className="font-mono font-semibold">
                {parent.account.code}
              </TableHead>
              <TableHead className="font-semibold">
                {parent.account.name}
              </TableHead>
              <TableHead className="text-right font-mono font-semibold">
                <div className="flex items-center justify-end gap-2">
                  <span>
                    <span className="text-muted-foreground">Total Saldo:</span> {formatCurrency(parent.account.balance)}
                  </span>
                  {isTableExpanded ? (
                    <ChevronUp className="h-4 w-4 shrink-0" />
                  ) : (
                    <ChevronDown className="h-4 w-4 shrink-0" />
                  )}
                </div>
              </TableHead>
            </TableRow>
          </TableHeader>
          {isTableExpanded && (
            <TableBody>
              {rows.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={3} className="h-16 text-center text-muted-foreground">
                    Belum ada sub-akun.
                  </TableCell>
                </TableRow>
              ) : (
              rows.map(({ node, depth, hasChildren }) => (
                <TableRow key={node.account.id} className="hover:bg-muted/50">
                  <TableCell className="w-50 font-mono text-sm">
                    <div style={{ paddingLeft: `${depth * 1.5}rem` }}>
                      {node.account.code}
                    </div>
                  </TableCell>
                  <TableCell>
                    <div
                      className="flex items-center gap-1"
                      style={{ paddingLeft: `${depth * 1.5}rem` }}
                    >
                      {hasChildren ? (
                        <button
                          onClick={() => onToggle(node.account.id)}
                          className="flex h-5 w-5 shrink-0 items-center justify-center rounded hover:bg-muted"
                        >
                          {expandedIds.has(node.account.id) ? (
                            <ChevronDown className="h-4 w-4" />
                          ) : (
                            <ChevronRight className="h-4 w-4" />
                          )}
                        </button>
                      ) : (
                        <span className="w-5 shrink-0" />
                      )}
                      <span className={hasChildren ? "font-medium" : ""}>
                        {node.account.name}
                      </span>
                    </div>
                  </TableCell>
                  <TableCell className="text-right font-mono">
                    {formatCurrency(node.account.balance)}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
          )}
        </Table>
      </div>
    </div>
  );
}

// ---- Main page ----
export default function AccountsPage() {
  const [tree, setTree] = useState<AccountTreeNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [expandedTableIds, setExpandedTableIds] = useState<Set<string>>(new Set());

  useEffect(() => {
    async function fetchTree() {
      try {
        setLoading(true);
        const res = await accountsApi.getTree();
        const data = Array.isArray(res)
          ? res
          : ((res as unknown as { data: AccountTreeNode[] }).data ?? []);
        setTree(data);
        // Expand all nodes and tables by default
        setExpandedIds(new Set(collectIds(data)));
        setExpandedTableIds(new Set(data.map((n) => n.account.id)));
      } catch (err) {
        setError(err instanceof Error ? err.message : "Failed to load accounts");
      } finally {
        setLoading(false);
      }
    }
    fetchTree();
  }, []);

  const toggleExpand = (id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const toggleTable = (id: string) => {
    setExpandedTableIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const filteredTree = filterTree(tree, search);

  return (
    <div className="space-y-6">
    <h1 className="text-2xl font-bold tracking-tight">Daftar Akun</h1>

      <div className="flex w-full items-center gap-4">
        <SearchBar
          containerClassName="flex-1"
          placeholder="Cari kode atau nama akun..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <Button>
          <Plus /> Tambah Akun Baru
        </Button>
      </div>

      {loading && (
        <div className="py-12 text-center text-muted-foreground">
          Memuat data akun...
        </div>
      )}

      {error && (
        <div className="py-12 text-center text-destructive">{error}</div>
      )}

      {!loading && !error && filteredTree.length === 0 && (
        <div className="py-12 text-center text-muted-foreground">
          {search ? "Tidak ada akun yang cocok." : "Belum ada data akun."}
        </div>
      )}

      {!loading &&
        !error &&
        filteredTree.map((parent) => (
          <AccountTable
            key={parent.account.id}
            parent={parent}
            expandedIds={expandedIds}
            onToggle={toggleExpand}
            isTableExpanded={expandedTableIds.has(parent.account.id)}
            onToggleTable={() => toggleTable(parent.account.id)}
          />
        ))}
    </div>
  );
}
