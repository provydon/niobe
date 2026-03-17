<script setup lang="ts">
import { Head, Link, router } from '@inertiajs/vue3';
import { ChevronDown, Loader2 } from 'lucide-vue-next';
import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import api from '@/lib/api';
import Heading from '@/components/Heading.vue';
import InputError from '@/components/InputError.vue';
import { normalizeNiobeAction, type NiobeAction } from '@/lib/niobe-actions';
import WaitressActionsInput from '@/components/WaitressActionsInput.vue';
import type { NiobeActionOption } from '@/lib/niobe-actions';
import { Button } from '@/components/ui/button';
import {
    Collapsible,
    CollapsibleContent,
    CollapsibleTrigger,
} from '@/components/ui/collapsible';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import AppLayout from '@/layouts/AppLayout.vue';
import { edit as waitressesEdit, index as waitressesIndex } from '@/routes/waitresses';
import { dashboard } from '@/routes';
import type { BreadcrumbItem } from '@/types';

interface MenuItemRow {
    id: number;
    name: string;
    category: string;
    unit_price: number | string;
    position: number;
}

interface Waitress {
    id: number;
    name: string;
    slug: string;
    context: string;
    menu_items?: MenuItemRow[];
    tools: { type: string; name: string; url?: string; target?: string }[];
    share_url: string;
    talk_url: string;
    tables_count?: number | null;
}

const props = defineProps<{
    waitressId: number;
}>();

const waitress = ref<Waitress | null>(null);
const actionTypes = ref<NiobeActionOption[]>([]);
const loading = ref(true);
const updateErrors = ref<Record<string, string>>({});
const processing = ref(false);

const name = ref('');
const tablesCount = ref(2);

const breadcrumbs = computed<BreadcrumbItem[]>(() => [
    { title: 'Dashboard', href: dashboard() },
    { title: 'My Waitresses', href: waitressesIndex() },
    { title: waitress.value ? `Edit: ${waitress.value.name}` : 'Edit', href: waitressesEdit(props.waitressId) },
]);

function emptyAction(): NiobeAction {
    const first = actionTypes.value[0]?.value ?? 'send_email';
    return { type: first, name: 'Place order', target: '' };
}

const actions = ref<NiobeAction[]>([emptyAction()]);

onMounted(async () => {
    try {
        const [waitressRes, typesRes] = await Promise.all([
            api.get<Waitress>(`/waitresses/${props.waitressId}`),
            api.get<{ actionTypes: NiobeActionOption[] }>('/waitresses/action-types'),
        ]);
        waitress.value = waitressRes.data;
        actionTypes.value = typesRes.data.actionTypes ?? [];
        name.value = waitress.value.name;
        tablesCount.value = waitress.value.tables_count ?? 2;
        if (Array.isArray(waitress.value.tools) && waitress.value.tools.length) {
            actions.value = waitress.value.tools.map((tool) => {
                const a = normalizeNiobeAction(tool);
                const validType = actionTypes.value.some((o) => o.value === a.type) ? a.type : actionTypes.value[0]?.value ?? 'send_email';
                return { ...a, type: validType };
            });
        } else {
            actions.value = [emptyAction()];
        }
        if ((waitress.value?.menu_items?.length ?? 0) === 0) {
            menuExtracting.value = true;
            menuPollInterval = setInterval(pollMenuItems, MENU_POLL_INTERVAL_MS);
            void pollMenuItems();
        }
    } finally {
        loading.value = false;
    }
});

const menuItems = computed(() => waitress.value?.menu_items ?? []);

const menuExtracting = ref(false);
let menuPollInterval: ReturnType<typeof setInterval> | null = null;
const MENU_POLL_INTERVAL_MS = 2500;
const MENU_POLL_MAX_ATTEMPTS = 48;
let menuPollAttempts = 0;

function stopMenuPolling() {
    if (menuPollInterval != null) {
        clearInterval(menuPollInterval);
        menuPollInterval = null;
    }
    menuExtracting.value = false;
}

async function pollMenuItems() {
    menuPollAttempts += 1;
    if (menuPollAttempts > MENU_POLL_MAX_ATTEMPTS) {
        stopMenuPolling();
        return;
    }
    try {
        const { data } = await api.get<MenuItemRow[]>(`/waitresses/${props.waitressId}/menu-items`);
        if (Array.isArray(data) && data.length > 0 && waitress.value) {
            stopMenuPolling();
            waitress.value = { ...waitress.value, menu_items: data };
        }
    } catch {
        // keep polling
    }
}

onBeforeUnmount(stopMenuPolling);

const editingMenuId = ref<number | null>(null);
const editMenuForm = ref({ name: '', category: '', unit_price: '' });
const newMenuItem = ref({ name: '', category: 'Other', unit_price: '' });
const menuItemErrors = ref<Record<string, string>>({});

function startEdit(item: MenuItemRow) {
    editingMenuId.value = item.id;
    editMenuForm.value = {
        name: item.name,
        category: item.category,
        unit_price: String(item.unit_price),
    };
    menuItemErrors.value = {};
}

function cancelEdit() {
    editingMenuId.value = null;
    menuItemErrors.value = {};
}

async function saveMenuItem() {
    if (editingMenuId.value == null) return;
    menuItemErrors.value = {};
    try {
        await api.put(`/waitresses/${props.waitressId}/menu-items/${editingMenuId.value}`, {
            name: editMenuForm.value.name,
            category: editMenuForm.value.category || 'Other',
            unit_price: editMenuForm.value.unit_price ? Number(editMenuForm.value.unit_price) : 0,
        });
        const { data } = await api.get<MenuItemRow[]>(`/waitresses/${props.waitressId}/menu-items`);
        if (waitress.value) waitress.value = { ...waitress.value, menu_items: data };
        editingMenuId.value = null;
    } catch (err: any) {
        if (err.response?.status === 422 && err.response?.data?.errors) {
            menuItemErrors.value = err.response.data.errors as Record<string, string>;
        }
    }
}

async function removeMenuItem(item: MenuItemRow) {
    if (!window.confirm('Remove this menu item?')) return;
    try {
        await api.delete(`/waitresses/${props.waitressId}/menu-items/${item.id}`);
        const { data } = await api.get<MenuItemRow[]>(`/waitresses/${props.waitressId}/menu-items`);
        if (waitress.value) waitress.value = { ...waitress.value, menu_items: data };
    } catch {
        // handled by interceptor
    }
}

async function addMenuItem() {
    menuItemErrors.value = {};
    try {
        await api.post(`/waitresses/${props.waitressId}/menu-items`, {
            name: newMenuItem.value.name,
            category: newMenuItem.value.category || 'Other',
            unit_price: newMenuItem.value.unit_price ? Number(newMenuItem.value.unit_price) : 0,
        });
        const { data } = await api.get<MenuItemRow[]>(`/waitresses/${props.waitressId}/menu-items`);
        if (waitress.value) waitress.value = { ...waitress.value, menu_items: data };
        newMenuItem.value = { name: '', category: 'Other', unit_price: '' };
    } catch (err: any) {
        if (err.response?.status === 422 && err.response?.data?.errors) {
            menuItemErrors.value = err.response.data.errors as Record<string, string>;
        }
    }
}

async function submitUpdate(e: Event) {
    e.preventDefault();
    updateErrors.value = {};
    processing.value = true;
    try {
        await api.put(`/waitresses/${props.waitressId}`, {
            name: name.value,
            tables_count: tablesCount.value,
            actions: actions.value,
        });
        router.visit(waitressesIndex(), { preserveScroll: false });
    } catch (err: any) {
        if (err.response?.status === 422 && err.response?.data?.errors) {
            updateErrors.value = (err.response.data.errors as Record<string, string[]>)?.reduce((acc, v, k) => ({ ...acc, [k]: v[0] }), {}) ?? {};
        }
    } finally {
        processing.value = false;
    }
}

async function deleteWaitress() {
    if (!window.confirm('Delete this waitress?')) return;
    try {
        await api.delete(`/waitresses/${props.waitressId}`);
        router.visit(waitressesIndex());
    } catch {
        // handled by interceptor
    }
}
</script>

<template>
    <AppLayout :breadcrumbs="breadcrumbs">
        <Head :title="waitress ? `Edit: ${waitress.name}` : 'Edit waitress'" />

        <div class="flex flex-1 flex-col gap-4 p-4 md:p-6">
            <div v-if="loading" class="flex justify-center py-12 text-muted-foreground">Loading…</div>
            <template v-else-if="waitress">
                <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                    <div>
                        <p class="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">Edit</p>
                        <Heading :title="waitress.name" variant="small" />
                    </div>
                    <a :href="waitress.share_url" target="_blank" rel="noopener" class="text-sm text-primary hover:underline">
                        View public page →
                    </a>
                </div>

                <div
                    v-if="menuExtracting"
                    class="flex items-center gap-3 rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-sm text-amber-700 dark:text-amber-400"
                >
                    <Loader2 class="h-5 w-5 shrink-0 animate-spin" />
                    <span>Menu is being extracted from your images. This page will update automatically when ready.</span>
                </div>

                <div class="rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-6 md:p-8">
                    <form class="space-y-6" @submit="submitUpdate">
                        <div class="space-y-2">
                            <Label for="name">Name</Label>
                            <Input id="name" v-model="name" required />
                            <InputError :message="updateErrors.name" />
                        </div>

                        <div class="space-y-2">
                            <Label for="tables_count">How many tables do you have?</Label>
                            <Input
                                id="tables_count"
                                v-model.number="tablesCount"
                                type="number"
                                min="0"
                                max="9999"
                                placeholder="e.g. 12"
                            />
                            <p class="text-xs text-muted-foreground">
                                Customers can say their table number (1 to this number), or use a link with ?table=5 for QR codes per table.
                            </p>
                            <InputError :message="updateErrors.tables_count" />
                        </div>

                        <WaitressActionsInput
                            v-model="actions"
                            :action-types="actionTypes"
                            :errors="updateErrors"
                            description="Where to send orders. A waitress can have multiple actions."
                            variant="edit"
                        />

                        <Collapsible :default-open="false" class="space-y-3">
                            <div class="flex items-center justify-between gap-2">
                                <CollapsibleTrigger
                                    class="flex flex-1 cursor-pointer items-center gap-2 rounded-lg border border-border bg-muted/30 px-3 py-2 text-left text-sm font-medium transition-colors hover:bg-muted/50 [&[data-state=open]>svg]:rotate-180"
                                    as-child
                                >
                                    <Button type="button" variant="ghost" class="flex flex-1 justify-between gap-2 px-0 font-medium">
                                        <span>Menu</span>
                                        <span class="text-muted-foreground">
                                            {{ menuItems.length }} {{ menuItems.length === 1 ? 'item' : 'items' }}
                                        </span>
                                        <ChevronDown class="h-4 w-4 shrink-0 transition-transform duration-200" />
                                    </Button>
                                </CollapsibleTrigger>
                            </div>
                            <CollapsibleContent class="space-y-3">
                                <p class="text-xs text-muted-foreground">
                                    Items the waitress can offer (name, category, unit price). Add, edit, or remove below.
                                </p>
                                <div v-if="menuItems.length" class="rounded-md border border-border bg-muted/30">
                                    <div
                                        v-for="item in menuItems"
                                        :key="item.id"
                                        class="flex flex-wrap items-center gap-2 border-b border-border px-2 py-2 last:border-0 sm:gap-3 sm:px-3"
                                    >
                                        <template v-if="editingMenuId === item.id">
                                            <input
                                                v-model="editMenuForm.name"
                                                type="text"
                                                placeholder="Name"
                                                class="h-9 min-w-0 flex-1 rounded-md border border-input bg-background px-2 text-sm"
                                            />
                                            <input
                                                v-model="editMenuForm.category"
                                                type="text"
                                                placeholder="Category"
                                                class="h-9 w-full rounded-md border border-input bg-background px-2 text-sm sm:w-28"
                                            />
                                            <input
                                                v-model="editMenuForm.unit_price"
                                                type="text"
                                                placeholder="Price"
                                                class="h-9 w-full rounded-md border border-input bg-background px-2 text-sm sm:w-24"
                                            />
                                            <div class="flex w-full gap-2 sm:w-auto">
                                                <Button type="button" size="sm" class="min-h-[44px] flex-1 sm:min-h-0 sm:flex-none" @click="saveMenuItem">Save</Button>
                                                <Button type="button" variant="ghost" size="sm" class="min-h-[44px] flex-1 sm:min-h-0 sm:flex-none" @click="cancelEdit">Cancel</Button>
                                            </div>
                                            <p v-if="menuItemErrors.name" class="w-full text-xs text-destructive">{{ menuItemErrors.name }}</p>
                                        </template>
                                        <template v-else>
                                            <span class="min-w-0 flex-1 truncate font-medium">{{ item.name }}</span>
                                            <span class="text-sm text-muted-foreground">{{ item.category }}</span>
                                            <span class="text-sm text-muted-foreground">{{ item.unit_price }}</span>
                                            <Button type="button" variant="ghost" size="sm" class="min-h-[44px] min-w-[44px] sm:min-h-0 sm:min-w-0" @click="startEdit(item)">Edit</Button>
                                            <Button type="button" variant="ghost" size="sm" class="min-h-[44px] min-w-[44px] text-destructive sm:min-h-0 sm:min-w-0" @click="removeMenuItem(item)">
                                                Remove
                                            </Button>
                                        </template>
                                    </div>
                                </div>
                                <div class="flex flex-wrap items-end gap-3 rounded-md border border-dashed border-border bg-muted/20 p-3">
                                    <div class="w-full min-w-0 flex-1 space-y-1 sm:w-auto sm:min-w-[120px]">
                                        <Label for="new-menu-name" class="text-xs">Name</Label>
                                        <Input id="new-menu-name" v-model="newMenuItem.name" placeholder="e.g. Cappuccino" class="h-9 w-full" />
                                    </div>
                                    <div class="w-full space-y-1 sm:w-28">
                                        <Label for="new-menu-category" class="text-xs">Category</Label>
                                        <Input id="new-menu-category" v-model="newMenuItem.category" placeholder="Other" class="h-9 w-full" />
                                    </div>
                                    <div class="w-full space-y-1 sm:w-24">
                                        <Label for="new-menu-price" class="text-xs">Unit price</Label>
                                        <Input
                                            id="new-menu-price"
                                            v-model="newMenuItem.unit_price"
                                            type="number"
                                            step="0.01"
                                            min="0"
                                            placeholder="0"
                                            class="h-9 w-full"
                                        />
                                    </div>
                                    <Button type="button" size="sm" class="min-h-[44px] w-full sm:min-h-0 sm:w-auto" :disabled="!newMenuItem.name.trim()" @click="addMenuItem">
                                        Add item
                                    </Button>
                                    <p v-if="menuItemErrors.name" class="w-full text-xs text-destructive">{{ menuItemErrors.name }}</p>
                                </div>
                            </CollapsibleContent>
                        </Collapsible>

                        <div class="mt-8 flex flex-col gap-3 sm:mt-10 sm:flex-row sm:flex-wrap sm:gap-3">
                            <Button type="submit" :disabled="processing" class="min-h-[44px] w-full rounded-xl px-5 py-3 sm:min-h-0 sm:w-auto">Update waitress</Button>
                            <Link :href="waitressesIndex()" class="w-full sm:w-auto">
                                <Button type="button" variant="secondary" class="min-h-[44px] w-full rounded-xl px-5 py-3 sm:min-h-0 sm:w-auto">Cancel</Button>
                            </Link>
                            <Button
                                type="button"
                                variant="ghost"
                                class="flex min-h-[44px] w-full items-center justify-center rounded-xl text-destructive hover:underline sm:ml-auto sm:min-h-0 sm:w-auto"
                                @click="deleteWaitress"
                            >
                                Delete waitress
                            </Button>
                        </div>
                    </form>
                </div>
            </template>
        </div>
    </AppLayout>
</template>

