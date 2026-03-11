import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuItem,
  SidebarMenuButton,
  SidebarTrigger,
} from "@/components/ui/sidebar";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import Link from "next/link";
import {
  LayoutDashboard,
  List,
  ArrowRightLeft,
  ArrowDownToLine,
  ArrowUpFromLine,
  FileText,
  Receipt,
  FileCheck,
  BookOpen,
  ClipboardList,
} from "lucide-react";

const menuGroups = [
  {
    label: "MENU",
    items: [
      { title: "Dashboard", icon: LayoutDashboard, href: "/" },
      { title: "Daftar Akun", icon: List, href: "/accounts" },
      { title: "Convert ke DJP", icon: ArrowRightLeft, href: "/convert-djp" },
    ],
  },
  {
    label: "KAS & BANK",
    items: [
      { title: "Penerimaan", icon: ArrowDownToLine, href: "/cash/receipts" },
      { title: "Pengeluaran", icon: ArrowUpFromLine, href: "/cash/payments" },
    ],
  },
  {
    label: "PENJUALAN",
    items: [
      { title: "Penagihan", icon: FileText, href: "/sales/invoices" },
      { title: "Faktur Pajak Penjualan", icon: Receipt, href: "/sales/tax-invoices" },
    ],
  },
  {
    label: "PEMBELIAN",
    items: [
      { title: "Faktur Pajak Pembelian", icon: Receipt, href: "/purchases/tax-invoices" },
    ],
  },
  {
    label: "PAJAK",
    items: [
      { title: "Review Faktur Pajak", icon: FileCheck, href: "/tax/review" },
    ],
  },
  {
    label: "BUKU BESAR",
    items: [
      { title: "Jurnal Umum", icon: BookOpen, href: "/ledger/journal" },
      { title: "Review Transaksi", icon: ClipboardList, href: "/ledger/review" },
    ],
  },
];

export function AppSidebar() {
  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <div className="flex items-center gap-2">
            <div className="group-data-[collapsible=icon]:hidden">
              <Select defaultValue="accounting">
                <SelectTrigger className="w-full">
                  <SelectValue placeholder="Select workspace" />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="accounting">Accounting & Tax</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <SidebarTrigger />
        </div>
      </SidebarHeader>
      <SidebarContent>
        {menuGroups.map((group) => (
          <SidebarGroup key={group.label}>
            <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {group.items.map((item) => (
                  <SidebarMenuItem key={item.title}>
                    <SidebarMenuButton asChild tooltip={item.title}>
                      <Link href={item.href}>
                        <item.icon />
                        <span>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
      <SidebarFooter>

      </SidebarFooter>
    </Sidebar>
  );
}