"use client";

import { customersApi, invoicesApi, paymentsApi } from "@/lib/api";
import { CreatePaymentRequest, Customer, Invoice, Payment, ReceivablesSummary } from "@/lib/types";
import { useCallback, useEffect, useMemo, useState } from "react";

export interface SalesController {
    loading: boolean;
    loadingSummary: boolean;
    loadingInvoices: boolean;
    error: string | null;
    payments: Payment[];
    filteredPayments: Payment[];
    invoices: Invoice[];
    filteredInvoices: Invoice[];
    search: string;
    summary: ReceivablesSummary | null;
    customers: Customer[];
    customerInvoices: Invoice[];

    fetchPayments: () => Promise<void>;
    fetchSummary: () => Promise<void>;
    fetchInvoices: () => Promise<void>;
    fetchCustomers: () => Promise<void>;
    fetchInvoicesByCustomer: (customerId: string) => Promise<void>;
    setSearch: (query: string) => void;
    recordPayment: (data: CreatePaymentRequest) => Promise<void>;
}

export function useSalesController(): SalesController {
    const [loading, setLoading] = useState(false);
    const [loadingSummary, setLoadingSummary] = useState(false);
    const [loadingInvoices, setLoadingInvoices] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [payments, setPayments] = useState<Payment[]>([]);
    const [invoices, setInvoices] = useState<Invoice[]>([]);
    const [summary, setSummary] = useState<ReceivablesSummary | null>(null);
    const [search, setSearch] = useState("");
    const [customers, setCustomers] = useState<Customer[]>([]);
    const [customerInvoices, setCustomerInvoices] = useState<Invoice[]>([]);

    const fetchPayments = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            const res = await paymentsApi.list();
            const data: Payment[] = Array.isArray(res)
                ? res
                : (res as unknown as { data: Payment[] }).data;
            setPayments(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to load payments");
        } finally {
            setLoading(false);
        }
    }, []);

    const fetchSummary = useCallback(async () => {
        try {
            setLoadingSummary(true);
            setError(null);
            const res = await paymentsApi.getSummary();
            const data: ReceivablesSummary =
                (res as unknown as { data: ReceivablesSummary }).data ?? res;
            setSummary(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to load summary");
        } finally {
            setLoadingSummary(false);
        }
    }, []);

    const fetchInvoices = useCallback(async () => {
        try {
            setLoadingInvoices(true);
            setError(null);
            const res = await invoicesApi.list({ status: ["unpaid", "partially_paid"] });
            const data: Invoice[] = Array.isArray(res)
                ? res
                : (res as unknown as { data: Invoice[] }).data ?? [];
            setInvoices(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to load invoices");
        } finally {
            setLoadingInvoices(false);
        }
    }, []);

    const recordPayment = useCallback(async (data: CreatePaymentRequest) => {
        try {
            setLoading(true);
            setError(null);
            await paymentsApi.record(data);
            await fetchPayments();
            await fetchSummary();
            await fetchInvoices();
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to record payment");
            throw err;
        } finally {
            setLoading(false);
        }
    }, [fetchPayments, fetchSummary, fetchInvoices]);

    useEffect(() => {
        fetchPayments();
        fetchSummary();
        fetchInvoices();
    }, [fetchPayments, fetchSummary, fetchInvoices]);

    const filteredPayments = useMemo(() => {
        const query = search.toLowerCase();
        return payments.filter(
            (p) =>
                p.payment_number.toLowerCase().includes(query) ||
                (p.customer_name ?? "").toLowerCase().includes(query)
        );
    }, [payments, search]);

    const filteredInvoices = useMemo(() => {
        const query = search.toLowerCase();
        return invoices.filter(
            (inv) =>
                inv.invoice_number.toLowerCase().includes(query) ||
                (inv.customer_name ?? "").toLowerCase().includes(query)
        );
    }, [invoices, search]);

    const fetchCustomers = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            const res = await customersApi.list();
            const data: Customer[] = Array.isArray(res)
                ? res
                : (res as unknown as { data: Customer[] }).data;
            setCustomers(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to load customers");
        } finally {
            setLoading(false);
        }
    }, []);

    const fetchInvoicesByCustomer = useCallback(async (customerId: string) => {
        if (!customerId) {
            setCustomerInvoices([]);
            return;
        }
        try {
            setLoadingInvoices(true);
            setError(null);
            const res = await invoicesApi.list({ customer_id: customerId });
            const data: Invoice[] = Array.isArray(res)
                ? res
                : (res as unknown as { data: Invoice[] }).data ?? [];
            setCustomerInvoices(data);
        } catch (err) {
            setError(err instanceof Error ? err.message : "Failed to load invoices");
            setCustomerInvoices([]);
        } finally {
            setLoadingInvoices(false);
        }
    }, []);

    return {
        loading,
        loadingSummary,
        loadingInvoices,
        error,
        payments,
        filteredPayments,
        invoices,
        filteredInvoices,
        search,
        summary,
        customers,
        customerInvoices,
        fetchPayments,
        fetchSummary,
        fetchInvoices,
        fetchCustomers,
        fetchInvoicesByCustomer,
        setSearch,
        recordPayment,
    };
}
