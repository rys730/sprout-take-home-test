"use client";

import { invoicesApi, journalsApi } from "@/lib/api";
import { CreateJournalRequest, Invoice, JournalEntry, JournalLine } from "@/lib/types";
import { useCallback, useEffect, useMemo, useState } from "react"

export interface JournalsController {
    loading: boolean;
    error: string | null;
    datas: JournalEntry[];
    filteredDatas: JournalEntry[];
    search: string;
    newJournalLines: Partial<JournalLine[]>;
    editEntry: JournalEntry | null;
    loadingEdit: boolean;

    fetchTable: () => Promise<void>;
    fetchById: (id: string) => Promise<void>;
    setSearch: (query: string) => void;
    setNewJournalLines: (data: Partial<JournalLine[]>) => void;
    checkDebitCreditSum: (entries: { account_id: string; debit: number; credit: number }[]) => boolean;
    saveDraft: (data: CreateJournalRequest) => Promise<void>;
    postJournal: (data: CreateJournalRequest) => Promise<void>;
    updateJournal: (id: string, data: CreateJournalRequest) => Promise<void>;
    updateAndPostJournal: (id: string, data: CreateJournalRequest) => Promise<void>;
    deleteJournal: (id: string) => Promise<void>;
    reverseJournal: (id: string, reason: string) => Promise<void>;
    invoices: Invoice[];
    loadingInvoices: boolean;
    fetchInvoices: () => Promise<void>;
}

export function useJournalsController(): JournalsController {
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [datas, setDatas] = useState<JournalEntry[]>([]);
    const [search, setSearch] = useState("")
    const [newJournalLines, setNewJournalLines] = useState<Partial<JournalLine[]>>([]);
    const [editEntry, setEditEntry] = useState<JournalEntry | null>(null);
    const [loadingEdit, setLoadingEdit] = useState(false);
    const [invoices, setInvoices] = useState<Invoice[]>([]);
    const [loadingInvoices, setLoadingInvoices] = useState(false);

    const fetchTable = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            const res = await journalsApi.list();
            const data: JournalEntry[] = Array.isArray(res) ? res : (res as unknown as {data: JournalEntry[]}).data;
            setDatas(data);
        } catch (error) {
            setError(error instanceof Error ? error.message : "Failed to load journals");
        } finally {
            setLoading(false);
        }
    }, [])
    
    useEffect (() => {
        fetchTable();
    }, [fetchTable])

    const filteredDatas = useMemo(() => {
        return datas.filter((data) => {
            const query = search.toLowerCase();
            return (
                data.entry_number.toLowerCase().includes(query)
            );
        });
    }, [datas, search]);

    const checkDebitCreditSum = (entries: { account_id: string; debit: number; credit: number }[]) => {
        const totalDebit = entries.reduce((sum, entry) => sum + entry.debit, 0);
        const totalCredit = entries.reduce((sum, entry) => sum + entry.credit, 0);
        return totalDebit === totalCredit;
    }

    const saveDraft = async (data: CreateJournalRequest) => {
        try {
            setLoading(true);
            setError(null);
            await journalsApi.create(data);
            await fetchTable();
        } catch (error) {
            setError(error instanceof Error ? error.message : "Failed to save draft");
        } finally {
            setLoading(false);
        }
    }

    const postJournal = async (data: CreateJournalRequest) => {
        try {
            setLoading(true);
            setError(null);
            await journalsApi.create({ ...data, status: "posted" });
            await fetchTable();
        } catch (error) {
            setError(error instanceof Error ? error.message : "Failed to post journal");
        } finally {
            setLoading(false);
        }
    }

    const fetchById = useCallback(async (id: string) => {
        try {
            setLoadingEdit(true);
            setError(null);
            const res = await journalsApi.getById(id);
            const entry: JournalEntry = (res as unknown as { data: JournalEntry }).data ?? res;
            setEditEntry(entry);
        } catch (error) {
            setError(error instanceof Error ? error.message : "Failed to load journal");
        } finally {
            setLoadingEdit(false);
        }
    }, []);

    const updateJournal = async (id: string, data: CreateJournalRequest) => {
        try {
            setLoading(true);
            setError(null);
            await journalsApi.update(id, { ...data, lines: data.lines });
            await fetchTable();
        } catch (error) {
            setError(error instanceof Error ? error.message : "Failed to update journal");
        } finally {
            setLoading(false);
        }
    }

    const updateAndPostJournal = async (id: string, data: CreateJournalRequest) => {
        try {
            setLoading(true);
            setError(null);
            await journalsApi.update(id, { ...data, lines: data.lines });
            await journalsApi.post(id);
            await fetchTable();
        } catch (error) {
            setError(error instanceof Error ? error.message : "Failed to post journal");
        } finally {
            setLoading(false);
        }
    }

    const deleteJournal = async (id: string) => {
        try {
            setLoading(true);
            setError(null);
            await journalsApi.delete(id);
            await fetchTable();
        } catch (error) {
            setError(error instanceof Error ? error.message : "Failed to delete journal");
        } finally {
            setLoading(false);
        }
    }

    const reverseJournal = async (id: string, reason: string) => {
        try {
            setLoading(true);
            setError(null);
            await journalsApi.reverse(id, { reason });
            await fetchTable();
        } catch (error) {
            setError(error instanceof Error ? error.message : "Failed to reverse journal");
        } finally {
            setLoading(false);
        }
    }

    const fetchInvoices = useCallback(async () => {
        try {
            setLoadingInvoices(true);
            const res = await invoicesApi.list();
            const data: Invoice[] = Array.isArray(res) ? res : (res as unknown as { data: Invoice[] }).data ?? [];
            setInvoices(data);
        } catch {
            // silently fail — invoice list is optional in journal form
        } finally {
            setLoadingInvoices(false);
        }
    }, []);

    return { 
        loading,
        error,
        datas,
        fetchTable,
        filteredDatas,
        search,
        setSearch,
        newJournalLines,
        setNewJournalLines,
        checkDebitCreditSum,
        saveDraft,
        postJournal,
        editEntry,
        loadingEdit,
        fetchById,
        updateJournal,
        updateAndPostJournal,
        deleteJournal,
        reverseJournal,
        invoices,
        loadingInvoices,
        fetchInvoices,
    };
}