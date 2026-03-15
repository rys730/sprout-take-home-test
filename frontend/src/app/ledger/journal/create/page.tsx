"use client";

import { useSearchParams } from "next/navigation";
import { useAccountsController } from "@/controllers/accounts-controller";
import { useJournalsController } from "@/controllers/journals-controller";
import { CreateJournalView } from "@/views/create-journal-view";

export default function CreateJournalPage() {
  const searchParams = useSearchParams();
  const editId = searchParams.get("id");
  const accountsController = useAccountsController();
  const journalsController = useJournalsController();
  return (
    <CreateJournalView
      accountsController={accountsController}
      journalsController={journalsController}
      editId={editId}
    />
  );
}