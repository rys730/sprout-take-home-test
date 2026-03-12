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
): { node: AccountTreeNode; depth: number; hasChildren: boolean; isLast: boolean; ancestorIsLast: boolean[] }[] {
  const rows: { node: AccountTreeNode; depth: number; hasChildren: boolean; isLast: boolean; ancestorIsLast: boolean[] }[] =
    [];

  function walk(list: AccountTreeNode[], depth: number, ancestors: boolean[]) {
    for (let i = 0; i < list.length; i++) {
      const n = list[i];
      const hasChildren = !!n.children && n.children.length > 0;
      const isLast = i === list.length - 1;
      rows.push({ node: n, depth, hasChildren, isLast, ancestorIsLast: [...ancestors] });
      if (hasChildren && expandedIds.has(n.account.id)) {
        walk(n.children!, depth + 1, [...ancestors, isLast]);
      }
    }
  }

  walk(children, startDepth, []);
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

  // Create dialog
  isCreateOpen: boolean;
  createForm: CreateAccountForm;
  isCreating: boolean;
  createError: string | null;

  // Edit dialog
  isEditOpen: boolean;
  editForm: CreateAccountForm;
  editAccountId: string | null;
  isEditing: boolean;
  editError: string | null;

  // Delete confirmation
  isDeleteOpen: boolean;
  deleteTarget: Account | null;
  isDeleting: boolean;
  deleteError: string | null;

  setSearch: (query: string) => void;
  toggleExpand: (id: string) => void;
  toggleTable: (id: string) => void;
  refresh: () => void;

  setCreateOpen: (open: boolean) => void;
  setCreateField: <K extends keyof CreateAccountForm>(key: K, value: CreateAccountForm[K]) => void;
  submitCreate: () => Promise<void>;

  openEdit: (account: Account) => void;
  setEditOpen: (open: boolean) => void;
  setEditField: <K extends keyof CreateAccountForm>(key: K, value: CreateAccountForm[K]) => void;
  submitEdit: () => Promise<void>;

  openDelete: (account: Account) => void;
  setDeleteOpen: (open: boolean) => void;
  confirmDelete: () => Promise<void>;
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

  // ---- Edit dialog state ----
  const [isEditOpen, setEditOpen] = useState(false);
  const [editForm, setEditForm] = useState<CreateAccountForm>(emptyForm);
  const [editAccountId, setEditAccountId] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [editError, setEditError] = useState<string | null>(null);

  // ---- Delete confirmation state ----
  const [isDeleteOpen, setDeleteOpen] = useState(false);
  const [deleteTarget, setDeleteTarget] = useState<Account | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [deleteError, setDeleteError] = useState<string | null>(null);

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

  // ---- Edit dialog handlers ----

  const openEdit = useCallback((account: Account) => {
    setEditAccountId(account.id);
    setEditForm({
      name: account.name,
      parent_id: account.parent_id ?? "",
      code: account.code,
      starting_balance: String(account.balance),
    });
    setEditError(null);
    setEditOpen(true);
  }, []);

  const handleSetEditOpen = useCallback((open: boolean) => {
    setEditOpen(open);
    if (!open) {
      setEditForm(emptyForm);
      setEditAccountId(null);
      setEditError(null);
    }
  }, []);

  const setEditFieldCb = useCallback(
    <K extends keyof CreateAccountForm>(key: K, value: CreateAccountForm[K]) => {
      setEditForm((prev) => ({ ...prev, [key]: value }));
    },
    []
  );

  const submitEdit = useCallback(async () => {
    if (!editAccountId) return;
    setIsEditing(true);
    setEditError(null);
    try {
      await accountsApi.update(editAccountId, {
        code: editForm.code,
        name: editForm.name,
        parent_id: editForm.parent_id || undefined,
        balance: editForm.starting_balance ? Number(editForm.starting_balance) : undefined,
      });
      setEditOpen(false);
      setEditForm(emptyForm);
      setEditAccountId(null);
      await fetchTree();
    } catch (err) {
      setEditError(err instanceof Error ? err.message : "Gagal mengubah akun");
    } finally {
      setIsEditing(false);
    }
  }, [editAccountId, editForm, fetchTree]);

  // ---- Delete handlers ----

  const openDelete = useCallback((account: Account) => {
    setDeleteTarget(account);
    setDeleteError(null);
    setDeleteOpen(true);
  }, []);

  const handleSetDeleteOpen = useCallback((open: boolean) => {
    setDeleteOpen(open);
    if (!open) {
      setDeleteTarget(null);
      setDeleteError(null);
    }
  }, []);

  const confirmDelete = useCallback(async () => {
    if (!deleteTarget) return;
    setIsDeleting(true);
    setDeleteError(null);
    try {
      await accountsApi.delete(deleteTarget.id);
      setDeleteOpen(false);
      setDeleteTarget(null);
      await fetchTree();
    } catch (err) {
      setDeleteError(err instanceof Error ? err.message : "Gagal menghapus akun");
    } finally {
      setIsDeleting(false);
    }
  }, [deleteTarget, fetchTree]);

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

    isEditOpen,
    editForm,
    editAccountId,
    isEditing,
    editError,
    openEdit,
    setEditOpen: handleSetEditOpen,
    setEditField: setEditFieldCb,
    submitEdit,

    isDeleteOpen,
    deleteTarget,
    isDeleting,
    deleteError,
    openDelete,
    setDeleteOpen: handleSetDeleteOpen,
    confirmDelete,
  };
}
