"use client";

import { useSalesController } from "@/controllers/sales-controller";
import { SalesView } from "@/views/sales-view";

export default function SalesInvoicesPage() {
  const salesController = useSalesController();
  return (
    <SalesView salesController={salesController} />
  );
}
