"use client";

import { useEffect, useState, useCallback, useMemo } from "react";
import { accountsApi } from "@/lib/api/accounts";
import type { Account, AccountTreeNode, CreateAccountRequest } from "@/lib/types";

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

export function flattenChildren(
  children: AccountTreeNode[],
  expandedIds: Set<string>,
  startDepth: number = 0
): { node: AccountTreeNode; depth: number; hasChildren: boolean }[] {
  const rows: { node: AccountTreeNode; depth: number; hasChildren: boolean }[] =
    [];

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
          children:
            filteredChildren.length > 0 ? filteredChildren : n.children,
        });
      }
    }
    return result;
  }

  return filter(nodes);
}


export interface CreateAccountForm {
  name: string;
  parent_id: string;
  code: string;
  starting_balance: string;
}

const emptyForm: CreateAccountForm = {
  name: "",
  parent_id: "",
  code: "",
  starting_balance: "",
};

export interface AccountsController {
  tree: AccountTreeNode[];
  filteredTree: AccountTreeNode[];
  flatAccounts: Account[];
  loading: boolean;
  error: string | null;
  search: string;
  expandedIds: Set<string>;
  expandedTableIds: Set<string>;

  isCreateOpen: boolean;
  createForm: CreateAccountForm;
  isCreating: boolean;
  createError: string | null;

  setSearch: (query: string) => void;
  toggleExpand: (id: string) => void;
  toggleTable: (id: string) => void;
  refresh: () => void;
  setCreateOpen: (open: boolean) => void;
  setCreateField: <K extends keyof CreateAccountForm>(key: K, value: CreateAccountForm[K]) => void;
  submitCreate: () => Promise<void>;
}

export function useAccountsController(): AccountsController {
  const [tree, setTree] = useState<AccountTreeNode[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [expandedIds, setExpandedIds] = useState<Set<string>>(new Set());
  const [expandedTableIds, setExpandedTableIds] = useState<Set<string>>(
    new Set()
  );

  // ---- Create dialog state ----
  const [isCreateOpen, setCreateOpen] = useState(false);
  const [createForm, setCreateForm] = useState<CreateAccountForm>(emptyForm);
  const [isCreating, setIsCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  const fetchTree = useCallback(async () => {
    try {
      setLoading(true);
      setError(null);
      const res = await accountsApi.getTree();
      const data = Array.isArray(res)
        ? res
        : ((res as unknown as { data: AccountTreeNode[] }).data ?? []);
      setTree(data);
      // Expand all nodes and tables by default
      setExpandedIds(new Set(collectIds(data)));
      setExpandedTableIds(new Set(data.map((n) => n.account.id)));
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Failed to load accounts"
      );
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchTree();
  }, [fetchTree]);

  const toggleExpand = useCallback((id: string) => {
    setExpandedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const toggleTable = useCallback((id: string) => {
    setExpandedTableIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }, []);

  const filteredTree = useMemo(
    () => filterTree(tree, search),
    [tree, search]
  );

  // ---- Flat list of all accounts (for parent dropdown) ----
  const flatAccounts = useMemo<Account[]>(() => {
    const result: Account[] = [];
    function walk(nodes: AccountTreeNode[]) {
      for (const n of nodes) {
        result.push(n.account);
        if (n.children) walk(n.children);
      }
    }
    walk(tree);
    return result;
  }, [tree]);

  // ---- Create dialog handlers ----

  const handleSetCreateOpen = useCallback(
    (open: boolean) => {
      setCreateOpen(open);
      if (!open) {
        setCreateForm(emptyForm);
        setCreateError(null);
      }
    },
    []
  );

  const setCreateField = useCallback(
    <K extends keyof CreateAccountForm>(key: K, value: CreateAccountForm[K]) => {
      setCreateForm((prev) => ({ ...prev, [key]: value }));
    },
    []
  );

  const submitCreate = useCallback(async () => {
    setIsCreating(true);
    setCreateError(null);
    try {
      const payload: CreateAccountRequest = {
        code: createForm.code,
        name: createForm.name,
        parent_id: createForm.parent_id,
        starting_balance: createForm.starting_balance
          ? Number(createForm.starting_balance)
          : undefined,
      };
      await accountsApi.create(payload);
      setCreateOpen(false);
      setCreateForm(emptyForm);
      await fetchTree();
    } catch (err) {
      setCreateError(
        err instanceof Error ? err.message : "Gagal membuat akun"
      );
    } finally {
      setIsCreating(false);
    }
  }, [createForm, fetchTree]);

  return {
    tree,
    filteredTree,
    flatAccounts,
    loading,
    error,
    search,
    expandedIds,
    expandedTableIds,
    isCreateOpen,
    createForm,
    isCreating,
    createError,
    setSearch,
    toggleExpand,
    toggleTable,
    refresh: fetchTree,
    setCreateOpen: handleSetCreateOpen,
    setCreateField,
    submitCreate,
  };
}
