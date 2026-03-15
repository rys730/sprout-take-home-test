import { SearchBar } from "@/components/search-bar";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Table,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { JournalsController } from "@/controllers/journals-controller";
import { JournalEntry } from "@/lib/types";
import { Plus, RotateCcw, TriangleAlert } from "lucide-react";
import Link from "next/link";
import { useState } from "react";

function JournalsTable({
  datas,
  onReversePosted,
}: {
  datas: JournalEntry[];
  onReversePosted: (entry: JournalEntry) => void;
}) {
  return (
    <div>
      <div className="rounded-md border">
        <Table>
          <TableHeader className="bg-muted">
            <TableRow>
              <TableHead>Tanggal</TableHead>
              <TableHead>Nomor</TableHead>
              <TableHead>Deskripsi</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>Action</TableHead>
            </TableRow>
          </TableHeader>
          {datas.map((data) => {
            return (
              <TableRow key={data.id} className="bg-white hover:bg-muted/50">
                <TableCell>
                  {new Date(data.created_at).toLocaleDateString("id-ID", {
                    day: "2-digit",
                    month: "2-digit",
                    year: "numeric",
                  })}
                </TableCell>
                <TableCell>{data.entry_number}</TableCell>
                <TableCell>{data.description}</TableCell>
                <TableCell>
                  {data.status === "posted" ? (
                    <span className="text-muted-foreground">{data.status}</span>
                  ) : (
                    <span className="inline-block rounded-full border border-muted-foreground/30 bg-muted px-3 py-0.5 text-xs text-muted-foreground">
                      {data.status}
                    </span>
                  )}
                </TableCell>
                <TableCell>
                  {data.status === "draft" ? (
                    <Link
                      href={{
                        pathname: "/ledger/journal/create",
                        query: { id: data.id },
                      }}
                    >
                      <button
                        className="rounded p-1 hover:bg-muted"
                        title="Edit draft"
                      >
                        <RotateCcw className="h-4 w-4 text-black" />
                      </button>
                    </Link>
                  ) : data.status === "posted" ? (
                    <button
                      className="rounded p-1 hover:bg-orange-100"
                      title="Balik jurnal"
                      onClick={() => onReversePosted(data)}
                    >
                      <RotateCcw className="h-4 w-4 text-orange-500" />
                    </button>
                  ) : (
                    <button
                      className="rounded p-1 hover:bg-muted"
                      title="Reload"
                      disabled
                    >
                      <RotateCcw className="h-4 w-4 text-muted-foreground" />
                    </button>
                  )}
                </TableCell>
              </TableRow>
            );
          })}
        </Table>
      </div>
    </div>
  );
}

export function JournalsVliew(ctrl: JournalsController) {
  const [reverseTarget, setReverseTarget] = useState<JournalEntry | null>(null);
  const [reverseReason, setReverseReason] = useState("");

  const handleConfirmReverse = async () => {
    if (!reverseTarget) return;
    await ctrl.reverseJournal(reverseTarget.id, reverseReason);
    setReverseTarget(null);
    setReverseReason("");
  };

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold tracking-tight">Jurnal Umum</h1>
      <div className="flex w-full items-center gap-4">
        <SearchBar
          containerClassName="flex-1"
          placeholder="Cari akun..."
          value={ctrl.search}
          onChange={(e) => ctrl.setSearch(e.target.value)}
        />
        <Button asChild>
          <Link href="/ledger/journal/create">
            <Plus />
            Tambah Jurnal Baru
          </Link>
        </Button>
      </div>
      {!ctrl.loading && !ctrl.error ? (
        <JournalsTable
          datas={ctrl.filteredDatas}
          onReversePosted={setReverseTarget}
        />
      ) : (
        <></>
      )}

      <Dialog
        open={!!reverseTarget}
        onOpenChange={(open) => {
          if (!open) {
            setReverseTarget(null);
            setReverseReason("");
          }
        }}
      >
        <DialogContent className="sm:max-w-sm">
          <div className="flex flex-col items-center gap-4">
            <div className="flex h-16 w-16 items-center justify-center text-orange-500">
              <TriangleAlert className="h-full w-full" />
            </div>
            <span className="text-lg font-bold">Balik Jurnal</span>
            <span className="text-center text-sm text-muted-foreground">
              Jurnal{" "}
              <span className="font-semibold text-foreground">
                {reverseTarget?.entry_number}
              </span>{" "}
              akan dibalik. Masukkan alasan pembalikan.
            </span>
            <div className="w-full grid gap-1.5">
              <Label htmlFor="reverse-reason">Alasan</Label>
              <Input
                id="reverse-reason"
                placeholder="Contoh: Kesalahan pencatatan"
                value={reverseReason}
                onChange={(e) => setReverseReason(e.target.value)}
              />
            </div>
            <Button
              className="w-full"
              variant="destructive"
              onClick={handleConfirmReverse}
              disabled={ctrl.loading || reverseReason.trim() === ""}
            >
              {ctrl.loading ? "Memproses..." : "Ya, Balik Jurnal"}
            </Button>
            <Button
              className="w-full"
              variant="outline"
              onClick={() => {
                setReverseTarget(null);
                setReverseReason("");
              }}
            >
              Batal
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
