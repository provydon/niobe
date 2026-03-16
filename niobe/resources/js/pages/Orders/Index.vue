<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import AppLayout from '@/layouts/AppLayout.vue';
import { dashboard } from '@/routes';
import { edit as waitressEdit, index as waitressesIndex } from '@/routes/waitresses';
import type { BreadcrumbItem } from '@/types';

interface Waitress {
    id: number;
    name: string;
    slug: string;
}

interface Order {
    id: number;
    order_summary: string;
    sent_to: string;
    sent_at: string;
    table_number: string | null;
    customer_name: string | null;
    waitress: Waitress;
}

interface PaginatedOrders {
    data: Order[];
    links: { url: string | null; label: string; active: boolean }[];
}

defineProps<{
    orders: PaginatedOrders;
}>();

const ordersIndexUrl = '/orders';

const breadcrumbs: BreadcrumbItem[] = [
    { title: 'Dashboard', href: dashboard() },
    { title: 'Orders', href: ordersIndexUrl },
];

function formatSentAt(iso: string): string {
    const d = new Date(iso);
    return d.toLocaleString(undefined, {
        dateStyle: 'short',
        timeStyle: 'short',
    });
}

</script>

<template>
    <AppLayout :breadcrumbs="breadcrumbs">
        <Head title="Orders" />

        <div class="flex flex-1 flex-col gap-4 p-3 sm:p-4 md:p-6">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <p class="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">
                        Voice orders
                    </p>
                    <h1 class="mt-1 text-2xl font-semibold text-foreground">
                        Orders
                    </h1>
                    <p class="mt-1 text-sm text-muted-foreground">
                        Orders placed by customers via your Niobe waitresses. Each row shows when and where the order was sent.
                    </p>
                </div>
            </div>

            <div class="rounded-2xl border border-border bg-card shadow-2xl">
                <div class="overflow-x-auto">
                    <table class="w-full min-w-[520px] text-left text-sm">
                        <thead class="border-b border-border bg-muted/50">
                            <tr>
                                <th class="px-4 py-3 font-medium text-foreground sm:pl-6">
                                    Time
                                </th>
                                <th class="px-4 py-3 font-medium text-foreground">
                                    Waitress
                                </th>
                                <th class="px-4 py-3 font-medium text-foreground">
                                    Table
                                </th>
                                <th class="px-4 py-3 font-medium text-foreground">
                                    Customer
                                </th>
                                <th class="px-4 py-3 font-medium text-foreground">
                                    Action / Sent to
                                </th>
                                <th class="px-4 py-3 font-medium text-foreground pr-4 sm:pr-6">
                                    Order summary
                                </th>
                            </tr>
                        </thead>
                        <tbody class="divide-y divide-border">
                            <tr
                                v-for="order in orders.data"
                                :key="order.id"
                                class="align-top"
                            >
                                <td class="whitespace-nowrap px-4 py-3 text-muted-foreground sm:pl-6">
                                    {{ formatSentAt(order.sent_at) }}
                                </td>
                                <td class="px-4 py-3">
                                    <Link
                                        :href="order.waitress ? waitressEdit(order.waitress.id) : waitressesIndex()"
                                        class="text-primary hover:underline"
                                    >
                                        {{ order.waitress?.name ?? '—' }}
                                    </Link>
                                </td>
                                <td class="px-4 py-3 font-medium text-foreground">
                                    {{ order.table_number ?? '—' }}
                                </td>
                                <td class="px-4 py-3 text-muted-foreground">
                                    {{ order.customer_name ?? '—' }}
                                </td>
                                <td class="px-4 py-3 text-muted-foreground">
                                    {{ order.sent_to }}
                                </td>
                                <td class="max-w-md truncate px-4 py-3 pr-4 sm:pr-6" :title="order.order_summary">
                                    {{ order.order_summary }}
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div
                    v-if="!orders.data.length"
                    class="flex flex-col items-center justify-center gap-4 py-12 text-center"
                >
                    <p class="text-muted-foreground">
                        No orders yet. Orders will appear here when customers place them via your waitresses.
                    </p>
                    <Link
                        :href="waitressesIndex()"
                        class="text-sm text-primary hover:underline"
                    >
                        Manage waitresses
                    </Link>
                </div>
                <div
                    v-if="orders.links && orders.links.length > 3"
                    class="flex flex-wrap justify-center gap-2 border-t border-border px-4 py-4 sm:px-6"
                >
                    <template v-for="(link, i) in orders.links" :key="i">
                        <Link
                            v-if="link.url"
                            :href="link.url"
                            class="rounded-lg px-3 py-1.5 text-sm transition-colors"
                            :class="link.active ? 'bg-primary text-primary-foreground' : 'text-muted-foreground hover:bg-muted hover:text-foreground'"
                            preserve-scroll
                        >
                            <span v-html="link.label" />
                        </Link>
                        <span v-else class="px-3 py-1.5 text-sm text-muted-foreground" v-html="link.label" />
                    </template>
                </div>
            </div>
        </div>
    </AppLayout>
</template>
