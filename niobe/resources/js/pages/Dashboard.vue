<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import AppLayout from '@/layouts/AppLayout.vue';
import { dashboard as dashboardRoute } from '@/routes';
import { create as waitressesCreate } from '@/routes/waitresses';
import type { BreadcrumbItem } from '@/types';

defineProps<{
    totalOrders: number;
}>();

const breadcrumbs: BreadcrumbItem[] = [
    {
        title: 'Dashboard',
        href: dashboardRoute(),
    },
];
</script>

<template>
    <Head title="Dashboard" />

    <AppLayout :breadcrumbs="breadcrumbs">
        <div class="flex h-full flex-1 flex-col gap-4 overflow-x-auto p-4 md:p-6">
            <div class="grid grid-cols-1 gap-4 md:grid-cols-2">
                <div class="rounded-2xl border border-border bg-card p-8 shadow-2xl">
                    <p class="text-sm font-medium uppercase tracking-[0.2em] text-foreground/80">
                        Getting started
                    </p>
                    <h2 class="mt-2 text-3xl font-semibold text-foreground">
                        How to use Niobe
                    </h2>
                    <ul class="mt-5 space-y-3 text-base text-foreground/90">
                        <li class="flex gap-3">
                            <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" aria-hidden="true" />
                            <span>Create a waitress for your restaurant.</span>
                        </li>
                        <li class="flex gap-3">
                            <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" aria-hidden="true" />
                            <span>Share the link; customers talk to her by voice and place orders. </span>
                        </li>
                    </ul>
                    <Link
                        :href="waitressesCreate()"
                        class="mt-6 inline-flex w-fit items-center gap-2 rounded-xl bg-primary px-5 py-3 text-base font-medium text-primary-foreground transition-colors hover:opacity-90"
                    >
                        Create new waitress
                    </Link>
                </div>
                <div class="rounded-2xl border border-border bg-card p-8 shadow-2xl">
                    <p class="text-sm font-medium uppercase tracking-[0.2em] text-muted-foreground">
                        Orders
                    </p>
                    <h2 class="mt-2 text-xl font-semibold text-foreground">
                        Total orders
                    </h2>
                    <p class="mt-4 text-4xl font-semibold tabular-nums text-muted-foreground">
                        {{ totalOrders }}
                    </p>
                    <p class="mt-1 text-sm text-muted-foreground">
                        All time
                    </p>
                </div>
            </div>
        </div>
    </AppLayout>
</template>
