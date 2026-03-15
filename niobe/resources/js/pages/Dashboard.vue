<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import AppLayout from '@/layouts/AppLayout.vue';
import { create as waitressesCreate } from '@/routes/waitresses';
import { dashboard as dashboardRoute } from '@/routes';
import type { BreadcrumbItem } from '@/types';

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
            <div class="rounded-2xl border border-border bg-card p-8 shadow-2xl">
                <p class="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">
                    Getting started
                </p>
                <h2 class="mt-2 text-2xl font-semibold text-foreground">
                    How to use Niobe
                </h2>
                <ul class="mt-5 space-y-3 text-sm text-muted-foreground">
                    <li class="flex gap-3">
                        <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" aria-hidden="true" />
                        <span>Create a waitress for each place or role—e.g. a waitress with your menu, or a support rep with your company info.</span>
                    </li>
                    <li class="flex gap-3">
                        <span class="mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" aria-hidden="true" />
                        <span>Share one link; customers talk to her by voice.</span>
                    </li>
                </ul>
                <Link
                    :href="waitressesCreate()"
                    class="mt-6 inline-flex w-fit items-center gap-2 rounded-xl bg-primary px-5 py-3 text-sm font-medium text-primary-foreground transition-colors hover:opacity-90"
                >
                    Create new waitress
                </Link>
            </div>
        </div>
    </AppLayout>
</template>
