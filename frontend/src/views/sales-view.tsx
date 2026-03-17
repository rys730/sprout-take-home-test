"use client";

import { SearchBar } from "@/components/search-bar";
import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import { Input } from "@/components/ui/input";
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
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import type { SalesController } from "@/controllers/sales-controller";
import { useAccountsController } from "@/controllers/accounts-controller";
import type { Invoice } from "@/lib/types";
import { format } from "date-fns";
import { id } from "date-fns/locale";
import { BoxIcon, CalendarIcon, Plus } from "lucide-react";
import { useState } from "react";
import { Checkbox } from "@/components/ui/checkbox";

function daysOverdueLabel(dueDate: string): { label: string; className: string } {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const due = new Date(dueDate);
  due.setHours(0, 0, 0, 0);
  const diffMs = today.getTime() - due.getTime();
  const days = Math.round(diffMs / (1000 * 60 * 60 * 24));

  if (days === 0) {
    return { label: "Jatuh tempo hari ini", className: "text-muted-foreground" };
  } else if (days > 0) {
    return { label: `Terlambat ${days} hari`, className: "text-foreground" };
  } else {
    return { label: `Jatuh tempo ${Math.abs(days)} hari lagi`, className: "text-muted-foreground" };
  }
}

function InvoiceTable({ invoices }: { invoices: Invoice[] }) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader className="bg-muted">
          <TableRow>
            <TableHead className="font-semibold">Pelanggan</TableHead>
            <TableHead className="font-semibold">Nomor Faktur</TableHead>
            <TableHead className="font-semibold">Tanggal Jatuh Tempo</TableHead>
            <TableHead className="font-semibold">Umur</TableHead>
            <TableHead className="font-semibold">Sisa Tagihan</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {invoices.length === 0 ? (
            <TableRow>
              <TableCell colSpan={5} className="text-center text-muted-foreground py-8">
                Tidak ada data tagihan.
              </TableCell>
            </TableRow>
          ) : (
            invoices.map((inv) => {
              const remaining = inv.total_amount - inv.amount_paid;
              const { label, className } = daysOverdueLabel(inv.due_date);
              return (
                <TableRow key={inv.id} className="bg-white hover:bg-muted/50">
                  <TableCell>{inv.customer_name ?? inv.customer_id}</TableCell>
                  <TableCell className="text-muted-foreground">{inv.invoice_number}</TableCell>
                  <TableCell>
                    {new Date(inv.due_date).toLocaleDateString("id-ID", {
                      day: "2-digit",
                      month: "2-digit",
                      year: "numeric",
                    })}
                  </TableCell>
                  <TableCell className={className}>{label}</TableCell>
                  <TableCell className="font-medium">
                    Rp{remaining.toLocaleString("id-ID")}
                  </TableCell>
                </TableRow>
              );
            })
          )}
        </TableBody>
      </Table>
    </div>
  );
}

interface SalesViewProps {
  salesController: SalesController;
}

export function SalesView({ salesController }: SalesViewProps) {
  const {
    summary,
    loadingSummary,
    filteredInvoices,
    loadingInvoices,
    loading,
    search,
    setSearch,
    recordPayment,
    customers,
    fetchCustomers,
    customerInvoices,
    fetchInvoicesByCustomer,
  } = salesController;
  const accountsController = useAccountsController();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [paymentDate, setPaymentDate] = useState<Date>();
  const [customerId, setCustomerId] = useState("");
  const [depositAccountId, setDepositAccountId] = useState("");
  const [discountAccountId, setDiscountAccountId] = useState("");
  const [notes, setNotes] = useState("");
  const [selectedInvoiceIds, setSelectedInvoiceIds] = useState<Set<string>>(
    new Set(),
  );
  // per-invoice allocation amount input: invoiceId -> string amount
  const [allocationAmounts, setAllocationAmounts] = useState<
    Map<string, string>
  >(new Map());

  const resetForm = () => {
    setPaymentDate(undefined);
    setCustomerId("");
    setDepositAccountId("");
    setDiscountAccountId("");
    setNotes("");
    setSelectedInvoiceIds(new Set());
    setAllocationAmounts(new Map());
  };

  const toggleInvoice = (invoiceId: string, remaining: number) => {
    setSelectedInvoiceIds((prev) => {
      const next = new Set(prev);
      if (next.has(invoiceId)) {
        next.delete(invoiceId);
      } else {
        next.add(invoiceId);
        // pre-fill allocation with the full remaining amount
        setAllocationAmounts((am) => {
          const nm = new Map(am);
          if (!nm.has(invoiceId)) nm.set(invoiceId, remaining.toString());
          return nm;
        });
      }
      return next;
    });
  };

  const setAllocationAmount = (invoiceId: string, value: string) => {
    setAllocationAmounts((prev) => new Map(prev).set(invoiceId, value));
  };

  const selectedAllocations = Array.from(selectedInvoiceIds).map((invId) => ({
    invoice_id: invId,
    amount: parseFloat(allocationAmounts.get(invId) ?? "0"),
  }));

  // total payment amount is the sum of all selected allocations
  const totalAmount = selectedAllocations.reduce((sum, a) => sum + (isNaN(a.amount) ? 0 : a.amount), 0);

  const isFormValid =
    !!paymentDate &&
    customerId.trim() !== "" &&
    totalAmount > 0 &&
    depositAccountId !== "" &&
    selectedAllocations.length > 0 &&
    selectedAllocations.every((a) => a.amount > 0);

  const handleSubmit = async () => {
    if (!isFormValid || !paymentDate) return;
    try {
      await recordPayment({
        customer_id: customerId,
        payment_date: format(paymentDate, "yyyy-MM-dd"),
        amount: totalAmount,
        deposit_to_account_id: depositAccountId,
        notes: notes || undefined,
        allocations: selectedAllocations,
      });
      setDialogOpen(false);
      resetForm();
    } catch {
      // error is stored in controller
    }
  };

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Penagihan</h1>
      </div>
      <div className="flex w-full gap-5">
        <div className="flex flex-col bg-white w-1/2 p-5 rounded-md">
          <div className="flex items-center space-x-2">
            <BoxIcon />
            <span className="font-medium">Total Piutang</span>
          </div>
          <span className="font-bold mt-5">
            {loadingSummary
              ? "Memuat..."
              : `Rp${(summary?.total_outstanding ?? 0).toLocaleString("id-ID")}`}
          </span>
        </div>
        <div className="flex flex-col bg-white w-1/2 p-5 rounded-md">
          <div className="flex items-center space-x-2">
            <BoxIcon />
            <span className="font-medium">Total Jatuh Tempo</span>
          </div>
          <span className="font-bold mt-5">
            {loadingSummary
              ? "Memuat..."
              : `Rp${(summary?.total_overdue ?? 0).toLocaleString("id-ID")}`}
          </span>
        </div>
      </div>
      <div className="flex w-full items-center gap-4">
        <SearchBar
          containerClassName="flex-1"
          placeholder="Cari pembayaran..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <Button
          onClick={() => {
            setDialogOpen(true);
            fetchCustomers();
          }}
        >
          <Plus /> Catat Pembayaran
        </Button>
      </div>
      {loading ? (
        <div className="py-12 text-center text-muted-foreground">
          Memuat data...
        </div>
      ) : (
        <InvoiceTable invoices={filteredInvoices} />
      )}

      <Dialog
        open={dialogOpen}
        onOpenChange={(open) => {
          if (!open) {
            setDialogOpen(false);
            resetForm();
          }
        }}
      >
        <DialogContent className="w-full sm:max-w-2xl max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>Catat Pembayaran Pelanggan</DialogTitle>
          </DialogHeader>

          <div className="grid gap-4 py-2">
            {/* Customer ID */}
            <div className="grid gap-1.5">
              <Label>
                <span className="text-destructive">*</span>Nama Pelanggan
              </Label>
              <Select
                value={customerId}
                onValueChange={(val) => {
                  setCustomerId(val);
                  fetchInvoicesByCustomer(val);
                }}
              >
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Pilih pelanggan" />
                </SelectTrigger>
                <SelectContent>
                  {customers.map((c) => (
                    <SelectItem key={c.id} value={c.id}>
                      {c.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Payment Date */}
            <div className="flex gap-2">
              <div className="grid gap-1.5 w-1/3">
                <Label>
                  <span className="text-destructive">*</span> Tanggal Pembayaran
                </Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant="outline"
                      className={
                        "w-full justify-between text-left font-normal " +
                        (!paymentDate ? "text-muted-foreground" : "")
                      }
                    >
                      {paymentDate
                        ? format(paymentDate, "dd MMMM yyyy", { locale: id })
                        : "DD/MM/YYYY"}
                      <CalendarIcon className="h-4 w-4" />
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0" align="start">
                    <Calendar
                      mode="single"
                      selected={paymentDate}
                      onSelect={setPaymentDate}
                      initialFocus
                    />
                  </PopoverContent>
                </Popover>
              </div>

              {/* Deposit Account */}
              <div className="grid gap-1.5 w-1/3">
                <Label>
                  <span className="text-destructive">*</span> Setor ke Akun
                </Label>
                <Select
                  value={depositAccountId}
                  onValueChange={setDepositAccountId}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Pilih akun" />
                  </SelectTrigger>
                  <SelectContent>
                    {accountsController.flatAccounts.map((acc) => (
                      <SelectItem key={acc.id} value={acc.id}>
                        {acc.code} - {acc.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>

              {/* Discount Account */}
              <div className="grid gap-1.5 w-1/3">
                <Label>Diskon</Label>
                <Select
                  value={discountAccountId}
                  onValueChange={setDiscountAccountId}
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Pilih diskon (jika ada)" />
                  </SelectTrigger>
                  <SelectContent>
                    {accountsController.flatAccounts.map((acc) => (
                      <SelectItem key={acc.id} value={acc.id}>
                        {acc.code} - {acc.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            </div>
            <hr className="border-border" />
            <div className="flex flex-col gap-2">
              <Label>Alokasi Pembayaran</Label>
              {!customerId ? (
                <div className="flex flex-col font-bold justify-center items-center py-6">
                  <BoxIcon className="w-16 h-16 m-3 text-muted-foreground" />
                  <span>Alokasi Pembayaran Masih Kosong</span>
                  <span className="font-light text-muted-foreground">
                    Silahkan pilih pelanggan terlebih dahulu.
                  </span>
                </div>
              ) : loadingInvoices ? (
                <div className="py-6 text-center text-muted-foreground text-sm">
                  Memuat tagihan...
                </div>
              ) : customerInvoices.length === 0 ? (
                <div className="flex flex-col font-bold justify-center items-center py-6">
                  <BoxIcon className="w-16 h-16 m-3 text-muted-foreground" />
                  <span>Tidak ada tagihan</span>
                  <span className="font-light text-muted-foreground">
                    Pelanggan ini tidak memiliki tagihan yang belum dibayar.
                  </span>
                </div>
              ) : (
                <div className="rounded-md border">
                  <Table>
                    <TableHeader className="bg-muted">
                      <TableRow>
                        <TableHead className="w-10"></TableHead>
                        <TableHead className="font-semibold">Faktur</TableHead>
                        <TableHead className="font-semibold">
                          Sisa Tagihan
                        </TableHead>
                        <TableHead className="font-semibold">
                          Alokasi Pembayaran
                        </TableHead>
                      </TableRow>
                    </TableHeader>
                    <TableBody>
                      {customerInvoices.map((inv) => {
                        const remaining = inv.total_amount - inv.amount_paid;
                        const isChecked = selectedInvoiceIds.has(inv.id);
                        return (
                          <TableRow
                            key={inv.id}
                            className="bg-white hover:bg-muted/50"
                          >
                            <TableCell>
                              <Checkbox
                                checked={isChecked}
                                onCheckedChange={() =>
                                  toggleInvoice(inv.id, remaining)
                                }
                              />
                            </TableCell>
                            <TableCell className="font-medium">
                              {inv.invoice_number}
                            </TableCell>
                            <TableCell>
                              Rp{remaining.toLocaleString("id-ID")}
                            </TableCell>
                            <TableCell>
                              <div className="relative">
                                <span className="absolute left-3 top-1/2 -translate-y-1/2 text-sm text-muted-foreground">
                                  Rp.
                                </span>
                              <Input
                                min={0}
                                max={remaining}
                                placeholder="0"
                                disabled={!isChecked}
                                value={
                                  isChecked
                                    ? (allocationAmounts.get(inv.id) ?? "")
                                    : ""
                                }
                                onChange={(e) =>
                                  setAllocationAmount(inv.id, e.target.value)
                                }
                                className="w-36 pl-10"
                              />
                              </div>
                            </TableCell>
                          </TableRow>
                        );
                      })}
                    </TableBody>
                  </Table>
                </div>
              )}
            </div>
          </div>

          <DialogFooter className="w-full">
            <Button
              className="w-1/2"
              variant="outline"
              onClick={() => {
                setDialogOpen(false);
                resetForm();
              }}
            >
              Batal
            </Button>
            <Button
              className="w-1/2"
              onClick={handleSubmit}
              disabled={!isFormValid || loading}
            >
              {loading ? "Menyimpan..." : "Simpan"}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
