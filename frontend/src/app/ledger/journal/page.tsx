"use client";

import { useJournalsController } from "@/controllers/journals-controller";
import { JournalsVliew } from "@/views/journals-view";

export default function JournalPage() {
  const controller = useJournalsController();
  return <JournalsVliew {...controller} />
}