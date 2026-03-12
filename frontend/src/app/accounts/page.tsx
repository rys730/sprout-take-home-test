"use client";

import { useAccountsController } from "@/controllers/accounts-controller";
import { AccountsView } from "@/views/accounts-view";

export default function AccountsPage() {
  const controller = useAccountsController();
  return <AccountsView {...controller} />;
}

