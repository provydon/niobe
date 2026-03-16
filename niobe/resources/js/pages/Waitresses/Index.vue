<script setup lang="ts">
import { Head, Link, router, usePage } from '@inertiajs/vue3';
import { Loader2 } from 'lucide-vue-next';
import { onBeforeUnmount, onMounted, ref } from 'vue';
import AppLayout from '@/layouts/AppLayout.vue';
import { Button } from '@/components/ui/button';
import {
    Tooltip,
    TooltipContent,
    TooltipProvider,
    TooltipTrigger,
} from '@/components/ui/tooltip';
import api from '@/lib/api';
import { create, destroy, edit, index as waitressesIndex } from '@/routes/waitresses';
import { dashboard } from '@/routes';
import type { BreadcrumbItem } from '@/types';

interface Waitress {
    id: number;
    name: string;
    slug: string;
    context: string;
    share_url: string;
    talk_url: string;
    menu_items_count: number;
}

interface PaginatedWaitresses {
    data: Waitress[];
    links: { url: string | null; label: string; active: boolean }[];
}

const props = defineProps<{
    waitresses: PaginatedWaitresses;
}>();

const page = usePage();
const flash = (page.props as { flash?: { success?: string } }).flash;
const copyTooltipOpen = ref<number | null>(null);

const MENU_POLL_INTERVAL_MS = 2500;
const MENU_POLL_MAX_ATTEMPTS = 48;
let menuPollInterval: ReturnType<typeof setInterval> | null = null;
let menuPollAttempts = 0;

async function pollExtractingMenus() {
    const extracting = props.waitresses.data.filter((w) => (w.menu_items_count ?? 0) === 0);
    if (extracting.length === 0) return;
    menuPollAttempts += 1;
    if (menuPollAttempts > MENU_POLL_MAX_ATTEMPTS) {
        if (menuPollInterval != null) {
            clearInterval(menuPollInterval);
            menuPollInterval = null;
        }
        return;
    }
    for (const w of extracting) {
        try {
            const { data } = await api.get<unknown[]>(`/waitresses/${w.id}/menu-items`);
            if (Array.isArray(data) && data.length > 0) {
                if (menuPollInterval != null) {
                    clearInterval(menuPollInterval);
                    menuPollInterval = null;
                }
                router.reload();
                return;
            }
        } catch {
            // continue to next waitress
        }
    }
}

onMounted(() => {
    const hasExtracting = props.waitresses.data.some((w) => (w.menu_items_count ?? 0) === 0);
    if (hasExtracting) {
        menuPollInterval = setInterval(pollExtractingMenus, MENU_POLL_INTERVAL_MS);
        void pollExtractingMenus();
    }
});

onBeforeUnmount(() => {
    if (menuPollInterval != null) {
        clearInterval(menuPollInterval);
        menuPollInterval = null;
    }
});

const breadcrumbs: BreadcrumbItem[] = [
    { title: 'Dashboard', href: dashboard() },
    { title: 'My Waitresses', href: waitressesIndex() },
];

function copyUrl(url: string, waitressId: number) {
    navigator.clipboard.writeText(url);
    copyTooltipOpen.value = waitressId;
    setTimeout(() => {
        copyTooltipOpen.value = null;
    }, 1500);
}

function confirmDelete(event: MouseEvent) {
    if (!window.confirm('Delete this waitress?')) {
        event.preventDefault();
    }
}
</script>

<template>
    <AppLayout :breadcrumbs="breadcrumbs">
        <Head title="My Waitresses" />

        <div class="flex flex-1 flex-col gap-4 p-3 sm:p-4 md:p-6">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <p class="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">AI waitresses</p>
                    <h1 class="mt-1 text-2xl font-semibold text-foreground">My Waitresses</h1>
                </div>
                <Link :href="create()">
                    <Button class="rounded-xl px-5 py-3">Create waitress</Button>
                </Link>
            </div>

            <div v-if="flash?.success" class="rounded-xl border border-primary/30 bg-primary/10 p-4 text-sm text-primary">
                {{ flash.success }}
            </div>

            <div class="rounded-2xl border border-border bg-card shadow-2xl">
                <div class="p-4 sm:p-6">
                    <template v-if="waitresses.data.length">
                        <div
                            v-for="waitress in waitresses.data"
                            :key="waitress.id"
                            class="border-b border-border py-4 last:border-0 sm:py-5"
                        >
                            <h3 class="text-lg font-semibold text-foreground sm:text-xl">{{ waitress.name }}</h3>
                            <p class="mt-1 truncate text-sm text-muted-foreground" :title="waitress.context">
                                {{ waitress.context?.slice(0, 72) ?? '' }}{{ (waitress.context?.length ?? 0) > 72 ? '…' : '' }}
                            </p>
                            <div class="mt-1 flex items-center gap-2">
                                <span class="text-xs text-muted-foreground">Menu:</span>
                                <template v-if="(waitress.menu_items_count ?? 0) > 0">
                                    <span class="text-xs font-medium text-muted-foreground">{{ waitress.menu_items_count }} items</span>
                                </template>
                                <div
                                    v-else
                                    class="inline-flex items-center gap-1.5 rounded-full border border-amber-500/30 bg-amber-500/10 px-2.5 py-1 text-xs font-medium text-amber-700 dark:text-amber-400"
                                >
                                    <Loader2 class="h-3.5 w-3.5 shrink-0 animate-spin" aria-hidden="true" />
                                    <span>Extracting menu…</span>
                                </div>
                            </div>
                            <div class="mt-3 flex flex-wrap items-center gap-2 sm:gap-3">
                                <a
                                    :href="waitress.share_url"
                                    target="_blank"
                                    rel="noopener"
                                    class="min-h-[44px] min-w-[44px] flex items-center justify-center rounded-md text-sm text-primary hover:underline sm:min-h-0 sm:min-w-0 sm:justify-start"
                                >
                                    Open link
                                </a>
                                <TooltipProvider :delay-duration="0">
                                    <Tooltip
                                        :open="copyTooltipOpen === waitress.id"
                                        @update:open="(v) => !v && (copyTooltipOpen = null)"
                                    >
                                        <TooltipTrigger as-child>
                                            <Button
                                                variant="ghost"
                                                size="sm"
                                                @click="copyUrl(waitress.share_url, waitress.id)"
                                            >
                                                Copy
                                            </Button>
                                        </TooltipTrigger>
                                        <TooltipContent>
                                            {{ copyTooltipOpen === waitress.id ? 'Copied' : 'Copy' }}
                                        </TooltipContent>
                                    </Tooltip>
                                </TooltipProvider>
                                <Link :href="edit(waitress.id)" class="min-h-[44px] min-w-[44px] flex items-center justify-center rounded-md text-sm text-primary hover:underline sm:min-h-0 sm:min-w-0 sm:justify-start">Edit</Link>
                                <Link
                                    :href="destroy.url(waitress.id)"
                                    method="delete"
                                    as="button"
                                    class="min-h-[44px] min-w-[44px] flex items-center justify-center rounded-md text-sm text-destructive hover:underline sm:min-h-0 sm:min-w-0 sm:justify-start"
                                    preserve-scroll
                                    @click="confirmDelete"
                                >
                                    Delete
                                </Link>
                            </div>
                        </div>
                    </template>
                    <div v-else class="flex flex-col items-center justify-center gap-4 py-12 text-center">
                        <p class="text-foreground">
                            No waitresses yet.
                        </p>
                        <Link :href="create()">
                            <Button class="rounded-xl px-5 py-3">Create waitress</Button>
                        </Link>
                    </div>
                </div>
                <div v-if="waitresses.links && waitresses.links.length > 3" class="flex flex-wrap justify-center gap-2 border-t border-border px-4 py-4 sm:px-6">
                    <div v-for="(link, i) in waitresses.links" :key="i" class="inline">
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
                    </div>
                </div>
            </div>
        </div>
    </AppLayout>
</template>
