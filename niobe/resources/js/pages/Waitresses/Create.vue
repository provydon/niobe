<script setup lang="ts">
import { Head, Link, router } from '@inertiajs/vue3';
import { ArrowLeft, Upload } from 'lucide-vue-next';
import { onBeforeUnmount, ref } from 'vue';
import Heading from '@/components/Heading.vue';
import InputError from '@/components/InputError.vue';
import api from '@/lib/api';
import { createEmptyNiobeAction, type NiobeAction } from '@/lib/niobe-actions';
import { toUrl } from '@/lib/utils';
import WaitressActionsInput from '@/components/WaitressActionsInput.vue';
import type { NiobeActionOption } from '@/lib/niobe-actions';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Spinner } from '@/components/ui/spinner';
import AppLayout from '@/layouts/AppLayout.vue';
import {
    create as waitressesCreate,
    index as waitressesIndex,
    store as waitressesStore,
} from '@/routes/waitresses';
import { dashboard } from '@/routes';
import type { BreadcrumbItem } from '@/types';

const props = defineProps<{
    actionTypes: NiobeActionOption[];
}>();

const breadcrumbs: BreadcrumbItem[] = [
    { title: 'Dashboard', href: dashboard() },
    { title: 'My Waitresses', href: waitressesIndex() },
    { title: 'Create waitress', href: waitressesCreate.url() },
];

const actions = ref<NiobeAction[]>([createEmptyNiobeAction()]);
const errors = ref<Record<string, string>>({});
interface FileWithPreview {
    file: File;
    url: string;
}
const menuFileEntries = ref<FileWithPreview[]>([]);
const fileInput = ref<HTMLInputElement | null>(null);
const formRef = ref<HTMLFormElement | null>(null);
const processing = ref(false);
const submitError = ref<string | null>(null);
const isDragging = ref(false);

onBeforeUnmount(() => {
    menuFileEntries.value.forEach((e) => URL.revokeObjectURL(e.url));
});

function addFiles(files: FileList | null) {
    if (!files?.length) return;
    const newEntries: FileWithPreview[] = [];
    for (let i = 0; i < files.length; i++) {
        const f = files[i];
        if (!f?.type.startsWith('image/')) continue;
        newEntries.push({ file: f, url: URL.createObjectURL(f) });
    }
    if (newEntries.length) {
        menuFileEntries.value = [...menuFileEntries.value, ...newEntries];
    }
    if (fileInput.value) fileInput.value.value = '';
}

function openFilePicker() {
    fileInput.value?.click();
}

function onFileChange(e: Event) {
    const input = e.target as HTMLInputElement;
    addFiles(input.files ?? null);
}

function removeFile(index: number) {
    const entry = menuFileEntries.value[index];
    if (entry) URL.revokeObjectURL(entry.url);
    menuFileEntries.value.splice(index, 1);
}

function onDrop(e: DragEvent) {
    e.preventDefault();
    isDragging.value = false;
    addFiles(e.dataTransfer?.files ?? null);
}

function onDragOver(e: DragEvent) {
    e.preventDefault();
    isDragging.value = true;
}

function onDragLeave() {
    isDragging.value = false;
}

function normalizeErrors(errorBag: Record<string, string[] | string>): Record<string, string> {
    return Object.fromEntries(
        Object.entries(errorBag).map(([key, value]) => [
            key,
            Array.isArray(value) ? value[0] : value,
        ]),
    );
}

async function submit() {
    if (!formRef.value || menuFileEntries.value.length === 0) {
        return;
    }
    processing.value = true;
    submitError.value = null;
    errors.value = {};

    try {
        const formData = new FormData(formRef.value);
        menuFileEntries.value.forEach((e) => formData.append('menu_files[]', e.file));
        await api.post(toUrl(waitressesStore.url()), formData);
        router.visit(toUrl(waitressesIndex()));
    } catch (error: any) {
        if (error.response?.status === 422 && error.response?.data?.errors) {
            errors.value = normalizeErrors(error.response.data.errors);
            return;
        }
        submitError.value = error.response?.data?.message ?? 'Failed to create.';
    } finally {
        processing.value = false;
    }
}
</script>

<template>
    <AppLayout :breadcrumbs="breadcrumbs">
        <Head title="Create waitress" />

        <div class="flex flex-1 flex-col gap-4 p-4 md:p-6">
            <Link :href="waitressesIndex()" class="inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground">
                <ArrowLeft class="h-4 w-4 shrink-0" />
                Back to waitresses
            </Link>
            <div>
                <p class="text-xs font-medium uppercase tracking-[0.2em] text-muted-foreground">New</p>
                <Heading title="Create waitress" variant="small" />
            </div>

            <div class="rounded-2xl border border-border bg-card p-4 shadow-2xl sm:p-6 md:p-8">
                <form ref="formRef" class="space-y-6" @submit.prevent="submit">
                    <div class="space-y-2">
                        <Label for="name">Name</Label>
                        <Input
                            id="name"
                            name="name"
                            required
                            placeholder="e.g. Jays, Café Waitress"
                        />
                        <InputError :message="errors.name" />
                    </div>

                    <div class="space-y-2">
                        <Label>Menu images</Label>
                        <input
                            ref="fileInput"
                            type="file"
                            accept="image/jpeg,image/png,image/gif,image/webp"
                            multiple
                            class="hidden"
                            @change="onFileChange"
                        />
                        <div
                            class="upload-zone rounded-xl border-2 border-dashed border-primary/60 bg-primary/5 p-6 text-center transition-colors"
                            :class="[
                                isDragging ? 'border-primary bg-primary/10' : 'cursor-pointer hover:border-primary hover:bg-primary/10',
                            ]"
                            role="button"
                            tabindex="0"
                            @click="openFilePicker"
                            @drop="onDrop"
                            @dragover="onDragOver"
                            @dragleave="onDragLeave"
                            @keydown.enter="openFilePicker"
                            @keydown.space.prevent="openFilePicker"
                        >
                            <Upload class="mx-auto mb-2 h-8 w-8 text-muted-foreground" />
                            <p v-if="menuFileEntries.length === 0" class="text-sm font-medium text-foreground">
                                add images of your menu
                            </p>
                            <p v-else class="text-sm font-medium text-foreground">
                                {{ menuFileEntries.length === 1 ? menuFileEntries[0].file.name : `${menuFileEntries.length} images` }}
                            </p>
                            <div
                                v-if="menuFileEntries.length > 0"
                                class="mt-4 flex flex-wrap justify-center gap-2"
                            >
                                <div
                                    v-for="(entry, idx) in menuFileEntries"
                                    :key="entry.url"
                                    class="relative h-20 w-20 shrink-0 overflow-hidden rounded-lg border border-border bg-muted shadow-sm"
                                >
                                    <img
                                        :src="entry.url"
                                        :alt="entry.file.name"
                                        class="h-full w-full object-cover"
                                    />
                                    <button
                                        type="button"
                                        class="absolute right-0.5 top-0.5 flex h-5 w-5 items-center justify-center rounded-full bg-destructive/90 text-[10px] font-bold text-destructive-foreground shadow hover:bg-destructive"
                                        aria-label="Remove"
                                        @click.stop="removeFile(idx)"
                                    >
                                        ×
                                    </button>
                                </div>
                            </div>
                        </div>
                        <InputError :message="errors['menu_files']" />
                    </div>

                    <WaitressActionsInput
                        v-model="actions"
                        :action-types="actionTypes"
                        :errors="errors"
                        description="Where to send orders (e.g. email or webhook)"
                        variant="create"
                    />

                    <div class="space-y-2">
                        <Label for="tables_count">How many tables do you have?</Label>
                        <Input
                            id="tables_count"
                            name="tables_count"
                            type="number"
                            min="0"
                            max="9999"
                            placeholder="e.g. 12"
                        />
                        <p class="text-xs text-muted-foreground">
                            Used so customers can say their table number (1 to this number), or use a link with ?table=5 for QR codes per table.
                        </p>
                        <InputError :message="errors.tables_count" />
                    </div>

                    <p v-if="submitError" class="text-sm text-destructive">
                        {{ submitError }}
                    </p>

                    <div class="mt-10 flex justify-center">
                        <Button type="submit" :disabled="processing || menuFileEntries.length === 0" class="min-h-[44px] w-full rounded-xl px-8 py-3 sm:min-h-0 sm:w-auto sm:min-w-[200px]">
                            <Spinner v-if="processing" class="mr-2 h-4 w-4" />
                            Create
                        </Button>
                    </div>
                </form>
            </div>
        </div>
    </AppLayout>
</template>
