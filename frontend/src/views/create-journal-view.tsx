"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Label } from "@/components/ui/label";
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from "@/components/ui/popover";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Textarea } from "@/components/ui/textarea";
import { Input } from "@/components/ui/input";
import { ArrowLeft, CalendarIcon, Plus, Trash2 } from "lucide-react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { format } from "date-fns";
import { id } from "date-fns/locale";
import type { AccountsController } from "@/controllers/accounts-controller";
import type { JournalsController } from "@/controllers/journals-controller";
import type { CreateJournalLine } from "@/lib/types";

interface CreateJournalViewProps {
  accountsController: AccountsController;
  journalsController: JournalsController;
  editId?: string | null;
}

export function CreateJournalView({
  accountsController,
  journalsController,
  editId,
}: CreateJournalViewProps) {
  const router = useRouter();
  const [date, setDate] = useState<Date>();
  const [invoiceId, setInvoiceId] = useState("");
  const [description, setDescription] = useState("");
  const [rows, setRows] = useState([
    { accountId: "", debit: "", credit: "" },
  ]);

  const { fetchById, editEntry, invoices, fetchInvoices } = journalsController;

  // Load existing draft when editId is provided
  useEffect(() => {
    if (!editId) return;
    fetchById(editId);
  }, [editId, fetchById]);

  // Populate form when editEntry is loaded
  useEffect(() => {
    if (!editEntry) return;
    if (editEntry.date) setDate(new Date(editEntry.date));
    if (editEntry.description) setDescription(editEntry.description);
    if (editEntry.invoice_id) setInvoiceId(editEntry.invoice_id);
    if (editEntry.lines && editEntry.lines.length > 0) {
      setRows(
        editEntry.lines.map((line) => ({
          accountId: line.account_id,
          debit: line.debit ? String(line.debit) : "",
          credit: line.credit ? String(line.credit) : "",
        }))
      );
    }
  }, [editEntry]);

  // Fetch invoice list on mount
  useEffect(() => {
    fetchInvoices();
  }, [fetchInvoices]);

  const addRow = () => {
    setRows([...rows, { accountId: "", debit: "", credit: "" }]);
  };

  const updateRow = (
    index: number,
    field: "accountId" | "debit" | "credit",
    value: string
  ) => {
    setRows(
      rows.map((row, i) => (i === index ? { ...row, [field]: value } : row))
    );
  };

  const removeRow = (index: number) => {
    setRows(rows.filter((_, i) => i !== index));
  };

  const accounts = accountsController.flatAccounts;

  // Build journal lines from rows
  const journalLines: CreateJournalLine[] = rows.map((row) => ({
    account_id: row.accountId,
    debit: parseFloat(row.debit) || 0,
    credit: parseFloat(row.credit) || 0,
  }));

  const totalDebit = journalLines.reduce((sum, l) => sum + l.debit, 0);
  const totalCredit = journalLines.reduce((sum, l) => sum + l.credit, 0);
  const isBalanced = journalsController.checkDebitCreditSum(journalLines);

  const hasValidLines = rows.every(
    (r) => r.accountId !== "" && (r.debit !== "" || r.credit !== "")
  );
  const isFormValid =
    !!date &&
    description.trim() !== "" &&
    hasValidLines &&
    isBalanced &&
    totalDebit > 0;

  const buildRequest = (status?: string) => ({
    date: date ? format(date, "yyyy-MM-dd") : "",
    description,
    invoice_id: invoiceId || undefined,
    status,
    lines: journalLines,
  });

  const handleSaveDraft = async () => {
    if (editId) {
      await journalsController.updateJournal(editId, buildRequest("draft"));
    } else {
      await journalsController.saveDraft(buildRequest("draft"));
    }
    router.push("/ledger/journal");
  };

  const handlePostJournal = async () => {
    if (editId) {
      await journalsController.updateAndPostJournal(editId, buildRequest());
    } else {
      await journalsController.postJournal(buildRequest());
    }
    router.push("/ledger/journal");
  };

  const isEditing = !!editId;

  return (
    <div className="space-y-6 w-3/4">
      <div className="flex items-center gap-3 bg-white p-5 rounded-md">
        <Button variant="ghost" size="icon" asChild>
          <Link href="/ledger/journal">
            <ArrowLeft className="h-5 w-5" />
          </Link>
        </Button>
        <h1 className="text-2xl font-bold tracking-tight">
          {isEditing ? "Edit Jurnal" : "Tambah Jurnal Baru"}
        </h1>
      </div>

      {journalsController.loadingEdit ? (
        <div className="py-12 text-center text-muted-foreground">
          Memuat data jurnal...
        </div>
      ) : (
      <div className="grid gap-6 rounded-md bg-white p-6">
        {/* Tanggal Jatuh Tempo */}
        <div className="grid gap-2">
          <Label>
            <span className="text-destructive">*</span> Tanggal Jatuh Tempo
          </Label>
          <Popover>
            <PopoverTrigger asChild>
              <Button
                variant="outline"
                className={
                  "w-full justify-between text-left font-normal " +
                  (!date ? "text-muted-foreground" : "")
                }
              >
                {date
                  ? format(date, "dd MMMM yyyy", { locale: id })
                  : "DD/MM/YYYY"}
                <CalendarIcon className="h-4 w-4" />
              </Button>
            </PopoverTrigger>
            <PopoverContent className="w-auto p-0" align="start">
              <Calendar
                mode="single"
                selected={date}
                onSelect={setDate}
                initialFocus
              />
            </PopoverContent>
          </Popover>
        </div>

        {/* Pilih Invoice */}
        <div className="grid gap-2">
          <Label>Pilih Invoice</Label>
          <Select value={invoiceId} onValueChange={setInvoiceId}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder="Pilih invoice" />
            </SelectTrigger>
            <SelectContent>
              {invoices.map((inv) => (
                <SelectItem key={inv.id} value={inv.id}>
                  {inv.invoice_number}{inv.customer_name ? ` — ${inv.customer_name}` : ""}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>

        {/* Deskripsi */}
        <div className="grid gap-2">
          <Label>
            <span className="text-destructive">*</span> Deskripsi
          </Label>
          <Textarea
            placeholder="Contoh: biaya ATK bulan Oktober"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            rows={4}
          />
        </div>

        {/* Journal Lines */}
        <div>
          <div className="grid grid-cols-3 gap-4 mb-2">
            <Label>Akun</Label>
            <Label>Debit</Label>
            <Label>Kredit</Label>
          </div>
          <hr className="border-border" />
          {rows.map((row, index) => (
            <div key={index} className="flex items-center gap-4 mt-4">
              <div className="grid grid-cols-3 gap-4 flex-1">
                <Select
                  value={row.accountId}
                  onValueChange={(v) => updateRow(index, "accountId", v)}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Pilih akun" />
                  </SelectTrigger>
                  <SelectContent>
                    {accounts.map((acc) => (
                      <SelectItem key={acc.id} value={acc.id}>
                        {acc.code} - {acc.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                    Rp.
                  </span>
                  <Input
                    className="pl-10"
                    placeholder="0"
                    value={row.debit}
                    onChange={(e) => updateRow(index, "debit", e.target.value)}
                  />
                </div>
                <div className="relative">
                  <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                    Rp.
                  </span>
                  <Input
                    className="pl-10"
                    placeholder="0"
                    value={row.credit}
                    onChange={(e) => updateRow(index, "credit", e.target.value)}
                  />
                </div>
              </div>
              {rows.length > 1 && (
                <button
                  type="button"
                  onClick={() => removeRow(index)}
                  className="rounded p-1 hover:bg-destructive/10 cursor-pointer"
                  title="Hapus baris"
                >
                  <Trash2 className="h-4 w-4 text-destructive" />
                </button>
              )}
            </div>
          ))}
          <button
            type="button"
            onClick={addRow}
            className="ml-auto flex items-center my-2 text-green-600 cursor-pointer hover:text-green-700"
          >
            <Plus />
            <span className="text-sm font-medium text-green-600">
              Tambah Baris
            </span>
          </button>
          <hr className="border-border" />
          <div className="py-2">
            <div className="flex justify-end gap-[10%] py-2">
              <span className="text-sm font-medium text-muted-foreground">
                Total Debit:
              </span>
              <span className="font-bold">
                Rp{totalDebit.toLocaleString("id-ID")}
              </span>
            </div>
            <div className="flex justify-end gap-[10%]">
              <span className="text-sm font-medium text-muted-foreground">
                Total Kredit:
              </span>
              <span className="font-bold">
                Rp{totalCredit.toLocaleString("id-ID")}
              </span>
            </div>
          </div>
          <hr className="border-border" />
          <div className="py-2 flex justify-end gap-2">
            <Button
              className="ml-auto bg-white text-black border border-border hover:bg-muted"
              disabled={
                !date ||
                description.trim() === "" ||
                journalsController.loading
              }
              onClick={handleSaveDraft}
            >
              {journalsController.loading ? "Menyimpan..." : "Simpan Draft"}
            </Button>
            <Button
              disabled={!isFormValid || journalsController.loading}
              onClick={handlePostJournal}
            >
              {journalsController.loading ? "Memproses..." : "Tambah Jurnal"}
            </Button>
          </div>
        </div>
      </div>
      )}
    </div>
  );
}
