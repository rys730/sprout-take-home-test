"use client";

import { SearchBar } from "@/components/search-bar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
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
import { Plus, ChevronDown, ChevronUp, Pencil, Trash2, Lock, TriangleAlert, CircleCheck } from "lucide-react";
import type { AccountTreeNode, Account } from "@/lib/types";
import type { AccountsController } from "@/controllers/accounts-controller";
import { flattenChildren } from "@/controllers/accounts-controller";

// ---- Helpers ----

function formatCurrency(amount: number): string {
  return new Intl.NumberFormat("id-ID", {
    style: "currency",
    currency: "IDR",
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  }).format(amount);
}

// ---- Sub-table for a single parent account ----

function AccountTable({
  parent,
  isTableExpanded,
  onToggleTable,
  onEdit,
  onDelete,
}: {
  parent: AccountTreeNode;
  isTableExpanded: boolean;
  onToggleTable: () => void;
  onEdit: (account: Account) => void;
  onDelete: (account: Account) => void;
}) {
  const children = parent.children ?? [];
  // Flatten all descendants — child rows are not expandable/collapsible
  const allExpanded = new Set<string>();
  function collectIds(nodes: AccountTreeNode[]) {
    for (const n of nodes) {
      allExpanded.add(n.account.id);
      if (n.children) collectIds(n.children);
    }
  }
  collectIds(children);
  const rows = flattenChildren(children, allExpanded, 1);

  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow
              className="bg-muted/50 cursor-pointer"
              onClick={onToggleTable}
            >
              <TableHead className="font-mono font-semibold">
                {parent.account.code}
              </TableHead>
              <TableHead className="font-semibold">
                {parent.account.name}
              </TableHead>
              <TableHead className="text-right font-mono font-semibold">
                <div className="flex items-center justify-end gap-2">
                  <span>
                    <span className="text-muted-foreground">Total Saldo:</span>{" "}
                    {formatCurrency(parent.account.balance)}
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
                  <TableCell
                    colSpan={3}
                    className="h-16 text-center text-muted-foreground"
                  >
                    Belum ada sub-akun.
                  </TableCell>
                </TableRow>
              ) : (
                rows.map(({ node, hasChildren, isLast, ancestorIsLast }) => {
                  const acct = node.account;
                  const isProtected = acct.is_system || acct.is_control;

                  return (
                    <TableRow
                      key={acct.id}
                      className="hover:bg-muted/50"
                    >
                      <TableCell className="w-50 font-mono text-sm">
                        <div className="flex items-center">
                          {ancestorIsLast.map((_, i) => (
                            <span key={i} className="inline-block w-6" />
                          ))}
                          <span className="inline-block w-6 text-center text-muted-foreground/60 select-none">
                            {isLast ? "└─" : "├─"}
                          </span>
                          <span>{acct.code}</span>
                        </div>
                      </TableCell>
                      <TableCell>
                        <div className="flex items-center gap-1.5">
                          <span className={hasChildren ? "font-medium" : ""}>
                            {acct.name}
                          </span>
                          {isProtected ? (
                            <Lock className="h-3.5 w-3.5 text-muted-foreground/60" />
                          ) : (
                            <>
                              <button
                                onClick={() => onEdit(acct)}
                                className="rounded p-1 hover:bg-muted"
                                title="Edit akun"
                              >
                                <Pencil className="h-3.5 w-3.5 text-muted-foreground" />
                              </button>
                              <button
                                onClick={() => onDelete(acct)}
                                className="rounded p-1 hover:bg-destructive/10"
                                title="Hapus akun"
                              >
                                <Trash2 className="h-3.5 w-3.5 text-destructive" />
                              </button>
                            </>
                          )}
                        </div>
                      </TableCell>
                      <TableCell className="text-right font-mono">
                        {formatCurrency(acct.balance)}
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          )}
        </Table>
      </div>
    </div>
  );
}

// ---- Main View ----

export function AccountsView(ctrl: AccountsController) {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Daftar Akun</h1>

      <div className="flex w-full items-center gap-4">
        <SearchBar
          containerClassName="flex-1"
          placeholder="Cari kode atau nama akun..."
          value={ctrl.search}
          onChange={(e) => ctrl.setSearch(e.target.value)}
        />
        <Dialog open={ctrl.isCreateOpen} onOpenChange={ctrl.setCreateOpen}>
          <Button onClick={() => ctrl.setCreateOpen(true)}>
            <Plus /> Tambah Akun Baru
          </Button>

          <DialogContent className="sm:max-w-md">
            <DialogHeader>
              <DialogTitle>Tambah Akun Baru</DialogTitle>
            </DialogHeader>

            <div className="grid gap-4 py-4">
              {/* Nama Akun */}
              <div className="grid gap-2">
                <Label htmlFor="create-name"><span className="text-destructive">*</span> Nama Akun</Label>
                <Input
                  id="create-name"
                  placeholder="Contoh: Pemasukan"
                  value={ctrl.createForm.name}
                  onChange={(e) => ctrl.setCreateField("name", e.target.value)}
                />
              </div>

              {/* Akun Induk & Nomor Akun */}
              <div className="flex gap-4">
                <div className="grid flex-1 gap-2">
                  <Label htmlFor="create-parent"><span className="text-destructive">*</span> Akun Induk</Label>
                  <Select
                    value={ctrl.createForm.parent_id}
                    onValueChange={(v) => ctrl.setCreateField("parent_id", v)}
                  >
                    <SelectTrigger id="create-parent">
                      <SelectValue placeholder="Pilih akun induk" />
                    </SelectTrigger>
                    <SelectContent>
                      {ctrl.flatAccounts.map((a) => (
                        <SelectItem key={a.id} value={a.id}>
                          {a.code} — {a.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>

                <div className="grid flex-1 gap-2">
                  <Label htmlFor="create-code"><span className="text-destructive">*</span> Nomor Akun</Label>
                  <Input
                    id="create-code"
                    placeholder="Contoh: 120.000"
                    value={ctrl.createForm.code}
                    onChange={(e) => ctrl.setCreateField("code", e.target.value)}
                  />
                  {ctrl.checkCodeExists && (
                    <p className="text-xs text-destructive">Nomor akun sudah terdaftar</p>
                  )}
                </div>
              </div>

              {/* Saldo */}
              <div className="grid gap-2">
                <Label htmlFor="create-balance"><span className="text-destructive">*</span> Saldo Awal</Label>
                <div className="flex items-center gap-2">
                  <span className="text-sm font-medium">Rp.</span>
                  <Input
                    id="create-balance"
                    placeholder="0"
                    value={ctrl.createForm.starting_balance}
                    onChange={(e) =>
                      ctrl.setCreateField("starting_balance", e.target.value)
                    }
                  />
                </div>
              </div>

              {ctrl.createError && (
                <p className="text-sm text-destructive">{ctrl.createError}</p>
              )}
            </div>

            <DialogFooter>
              {(() => {
                const isValid =
                  ctrl.createForm.name.trim() !== "" &&
                  ctrl.createForm.parent_id !== "" &&
                  ctrl.createForm.code.trim() !== "" &&
                  ctrl.createForm.starting_balance.trim() !== "" &&
                  !ctrl.checkCodeExists;
                return (
                  <Button
                    onClick={ctrl.submitCreate}
                    disabled={ctrl.isCreating || !isValid}
                    className={ "w-full " +
                      (isValid
                        ? "bg-green-600 hover:bg-green-700 text-white"
                        : "bg-muted text-muted-foreground hover:bg-muted")
                    }
                  >
                    {ctrl.isCreating ? "Menyimpan..." : "Simpan"}
                  </Button>
                );
              })()}
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      {ctrl.loading && (
        <div className="py-12 text-center text-muted-foreground">
          Memuat data akun...
        </div>
      )}

      {ctrl.error && (
        <div className="py-12 text-center text-destructive">{ctrl.error}</div>
      )}

      {!ctrl.loading && !ctrl.error && ctrl.filteredTree.length === 0 && (
        <div className="py-12 text-center text-muted-foreground">
          {ctrl.search
            ? "Tidak ada akun yang cocok."
            : "Belum ada data akun."}
        </div>
      )}

      {!ctrl.loading &&
        !ctrl.error &&
        ctrl.filteredTree.map((parent) => (
          <AccountTable
            key={parent.account.id}
            parent={parent}
            isTableExpanded={ctrl.expandedTableIds.has(parent.account.id)}
            onToggleTable={() => ctrl.toggleTable(parent.account.id)}
            onEdit={ctrl.openEdit}
            onDelete={ctrl.openDelete}
          />
        ))}

      {/* ---- Edit Account Dialog ---- */}
      <Dialog open={ctrl.isEditOpen} onOpenChange={ctrl.setEditOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Edit Akun</DialogTitle>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            <div className="grid gap-2">
              <Label htmlFor="edit-name"><span className="text-destructive">*</span> Nama Akun</Label>
              <Input
                id="edit-name"
                placeholder="Contoh: Pemasukan"
                value={ctrl.editForm.name}
                onChange={(e) => ctrl.setEditField("name", e.target.value)}
              />
            </div>

            <div className="flex gap-4">
              <div className="grid flex-1 gap-2">
                <Label htmlFor="edit-parent"><span className="text-destructive">*</span> Akun Induk</Label>
                <Select
                  value={ctrl.editForm.parent_id}
                  onValueChange={(v) => ctrl.setEditField("parent_id", v)}
                >
                  <SelectTrigger id="edit-parent">
                    <SelectValue placeholder="Pilih akun induk" />
                  </SelectTrigger>
                  <SelectContent>
                    {ctrl.flatAccounts.map((a) => (
                      <SelectItem key={a.id} value={a.id}>
                        {a.code} — {a.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              <div className="grid flex-1 gap-2">
                <Label htmlFor="edit-code"><span className="text-destructive">*</span> Nomor Akun</Label>
                <Input
                  id="edit-code"
                  placeholder="Contoh: 120.000"
                  value={ctrl.editForm.code}
                  onChange={(e) => ctrl.setEditField("code", e.target.value)}
                />
                {ctrl.checkEditCodeExists && (
                  <p className="text-xs text-destructive">Nomor akun sudah terdaftar</p>
                )}
              </div>
            </div>

            <div className="grid gap-2">
              <Label htmlFor="edit-balance"><span className="text-destructive">*</span> Saldo</Label>
              <div className="flex items-center gap-2">
                <span className="text-sm font-medium">Rp.</span>
                <Input
                  id="edit-balance"
                  placeholder="0"
                  value={ctrl.editForm.starting_balance}
                  onChange={(e) =>
                    ctrl.setEditField("starting_balance", e.target.value)
                  }
                />
              </div>
            </div>

            {ctrl.editError && (
              <p className="text-sm text-destructive">{ctrl.editError}</p>
            )}
          </div>

          <DialogFooter>
            {(() => {
              const isValid =
                ctrl.editForm.name.trim() !== "" &&
                ctrl.editForm.code.trim() !== "" &&
                !ctrl.checkEditCodeExists;
              return (
                <Button
                  onClick={ctrl.submitEdit}
                  disabled={ctrl.isEditing || !isValid}
                  className={"w-full " +
                    (isValid
                      ? "bg-green-600 hover:bg-green-700 text-white"
                      : "bg-muted text-muted-foreground hover:bg-muted")
                  }
                >
                  {ctrl.isEditing ? "Menyimpan..." : "Simpan Perubahan"}
                </Button>
              );
            })()}
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* ---- Delete Confirmation Dialog ---- */}
      <Dialog open={ctrl.isDeleteOpen} onOpenChange={ctrl.setDeleteOpen}>
        <DialogContent className="sm:max-w-sm">
          <div className="flex flex-col items-center gap-4 py-4 text-center">
            {/* Warning icon */}
            <div className="flex h-24 w-24 items-center justify-center">
              <TriangleAlert className="h-full w-full" />
            </div>

            {/* Header */}
            <h3 className="text-lg font-semibold">Hapus Akun</h3>

            {/* Description */}
            <p className="text-sm text-muted-foreground">
              Apakah kamu yakin ingin menghapus akun ini?
            </p>

            {ctrl.deleteError && (
              <p className="text-sm text-destructive">{ctrl.deleteError}</p>
            )}

            {/* Delete button */}
            <Button
              variant="destructive"
              className="w-full"
              onClick={ctrl.confirmDelete}
              disabled={ctrl.isDeleting}
            >
              {ctrl.isDeleting ? "Menghapus..." : "Hapus"}
            </Button>

            {/* Cancel button */}
            <Button
              variant="outline"
              className="w-full"
              onClick={() => ctrl.setDeleteOpen(false)}
              disabled={ctrl.isDeleting}
            >
              Batal
            </Button>
          </div>
        </DialogContent>
      </Dialog>

      {/* ---- Success Dialog ---- */}
      <Dialog open={ctrl.isSuccessOpen} onOpenChange={ctrl.setSuccessOpen}>
        <DialogContent className="sm:max-w-sm">
          <div className="flex flex-col items-center gap-4 py-4 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-full bg-green-100">
              <CircleCheck className="h-full w-full text-green-600" />
            </div>

            <h3 className="text-lg font-semibold">Berhasil Tambah Akun Baru</h3>

            <p className="text-sm text-muted-foreground">
              {ctrl.successMessage}
            </p>

            <Button
              className="w-full bg-green-600 hover:bg-green-700 text-white"
              onClick={() => ctrl.setSuccessOpen(false)}
            >
              Kembali
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
