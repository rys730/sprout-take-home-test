import { SearchBar } from "@/components/search-bar"
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

export default function AccountsPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold tracking-tight">Daftar Akun</h1>
        <p className="text-muted-foreground">
          Kelola daftar akun (Chart of Accounts) perusahaan Anda.
        </p>
      </div>
      <div className="w-full flex items-center gap-4">
        <SearchBar containerClassName="flex-1" />
        <Button><Plus /> Tambah Akun Baru</Button>
      </div>
    </div>
  );
}
