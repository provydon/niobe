<script setup lang="ts">
import { Form, Head, Link, router, usePage } from '@inertiajs/vue3';
import { ChevronDown } from 'lucide-vue-next';
import { computed, ref } from 'vue';
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
import { destroy as waitressesDestroy, edit as waitressesEdit, index as waitressesIndex, update as waitressesUpdate } from '@/routes/waitresses';
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
    extracted_context?: unknown[] | null;
    menu_items?: MenuItemRow[];
    tools: { type: string; name: string; url?: string; target?: string }[];
    share_url: string;
    talk_url: string;
    tables_count?: number | null;
}

const props = defineProps<{
    waitress: Waitress;
    actionTypes: NiobeActionOption[];
}>();

const breadcrumbs: BreadcrumbItem[] = [
    { title: 'Dashboard', href: dashboard() },
    { title: 'My Waitresses', href: waitressesIndex() },
    { title: `Edit: ${props.waitress.name}`, href: waitressesEdit(props.waitress.id) },
];

function emptyAction(): NiobeAction {
    const first = props.actionTypes[0]?.value ?? 'send_email';
    return { type: first, name: 'Place order', target: '' };
}

const actions = ref<NiobeAction[]>(
    Array.isArray(props.waitress.tools) && props.waitress.tools.length
        ? props.waitress.tools.map((tool) => {
              const a = normalizeNiobeAction(tool);
              const validType = props.actionTypes.some((o) => o.value === a.type) ? a.type : props.actionTypes[0]?.value ?? 'send_email';
              return { ...a, type: validType };
          })
        : [emptyAction()]
);

const updateForm = computed(() => waitressesUpdate.form.patch(props.waitress.id));
const page = usePage();
const flash = (page.props as { flash?: { success?: string } }).flash;

function confirmDelete(event: MouseEvent) {
    if (!window.confirm('Delete this waitress?')) {
        event.preventDefault();
    }
}

const menuItems = computed(() => props.waitress.menu_items ?? []);

const editingMenuId = ref<number | null>(null);
const editMenuForm = ref({ name: '', category: '', unit_price: '' });
const newMenuItem = ref({ name: '', category: 'Other', unit_price: '' });
const menuItemErrors = ref<Record<string, string>>({});

function menuItemUrl(path: string) {
    return `/waitresses/${props.waitress.id}/menu-items${path}`;
}

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

function saveMenuItem() {
    if (editingMenuId.value == null) return;
    menuItemErrors.value = {};
    router.put(menuItemUrl(`/${editingMenuId.value}`), {
        name: editMenuForm.value.name,
        category: editMenuForm.value.category || 'Other',
        unit_price: editMenuForm.value.unit_price ? Number(editMenuForm.value.unit_price) : 0,
    }, {
        preserveScroll: true,
        onError: (errors) => {
            menuItemErrors.value = errors as Record<string, string>;
        },
        onSuccess: () => {
            editingMenuId.value = null;
        },
    });
}

function removeMenuItem(item: MenuItemRow) {
    if (!window.confirm('Remove this menu item?')) return;
    router.delete(menuItemUrl(`/${item.id}`), { preserveScroll: true });
}

function addMenuItem() {
    menuItemErrors.value = {};
    router.post(menuItemUrl(''), {
        name: newMenuItem.value.name,
        category: newMenuItem.value.category || 'Other',
        unit_price: newMenuItem.value.unit_price ? Number(newMenuItem.value.unit_price) : 0,
    }, {
        preserveScroll: true,
        onError: (errors) => {
            menuItemErrors.value = errors as Record<string, string>;
        },
        onSuccess: () => {
            newMenuItem.value = { name: '', category: 'Other', unit_price: '' };
        },
    });
}
</script>

<template>
    <AppLayout :breadcrumbs="breadcrumbs">
        <Head :title="`Edit: ${waitress.name}`" />

        <div class="flex flex-1 flex-col gap-4 p-4 md:p-6">
            <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
                <div>
                    <p class="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">Edit</p>
                    <Heading :title="`${waitress.name}`" variant="small" />
                </div>
                <a :href="waitress.share_url" target="_blank" rel="noopener" class="text-sm text-primary hover:underline">
                    View public page →
                </a>
            </div>

            <div v-if="flash?.success" class="rounded-xl border border-primary/30 bg-primary/10 p-4 text-sm text-primary">
                {{ flash.success }}
            </div>

            <div class="rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-6 md:p-8">
            <Form
                :action="updateForm.action"
                method="post"
                class="space-y-6"
                :force-form-data="true"
                v-slot="{ errors, processing }"
            >
                <input type="hidden" name="_method" value="PATCH" />

                <div class="space-y-2">
                    <Label for="name">Name</Label>
                    <Input id="name" name="name" :default-value="waitress.name" required />
                    <InputError :message="errors.name" />
                </div>

                <div class="space-y-2">
                    <Label for="tables_count">How many tables do you have?</Label>
                    <Input
                        id="tables_count"
                        name="tables_count"
                        type="number"
                        min="0"
                        max="9999"
                        placeholder="e.g. 12"
                        :default-value="waitress.tables_count ?? 2"
                    />
                    <p class="text-xs text-muted-foreground">
                        Customers can say their table number (1 to this number), or use a link with ?table=5 for QR codes per table.
                    </p>
                    <InputError :message="errors?.tables_count" />
                </div>

                <WaitressActionsInput
                    v-model="actions"
                    :action-types="actionTypes"
                    :errors="errors"
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
                                <Input
                                    id="new-menu-name"
                                    v-model="newMenuItem.name"
                                    placeholder="e.g. Cappuccino"
                                    class="h-9 w-full"
                                />
                            </div>
                            <div class="w-full space-y-1 sm:w-28">
                                <Label for="new-menu-category" class="text-xs">Category</Label>
                                <Input
                                    id="new-menu-category"
                                    v-model="newMenuItem.category"
                                    placeholder="Other"
                                    class="h-9 w-full"
                                />
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
                    <Link
                        :href="waitressesDestroy.url(waitress.id)"
                        method="delete"
                        as="button"
                        class="flex min-h-[44px] w-full items-center justify-center rounded-xl text-sm text-destructive hover:underline sm:ml-auto sm:min-h-0 sm:w-auto"
                        preserve-scroll
                        @click="confirmDelete"
                    >
                        Delete waitress
                    </Link>
                </div>
            </Form>
            </div>
        </div>
    </AppLayout>
</template>
