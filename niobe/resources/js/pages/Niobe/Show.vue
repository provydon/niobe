<script setup lang="ts">
import { Head, usePage } from '@inertiajs/vue3';

interface NiobeShowProps {
    niobe: {
        name: string;
        context: string;
        talk_url: string;
    };
}

const props = defineProps<NiobeShowProps>();

const page = usePage();
const appName = (page.props as { name?: string }).name ?? 'Niobe';
const talkHref = `${props.niobe.talk_url}${props.niobe.talk_url.includes('?') ? '&' : '?'}autostart=1`;

const rawContext = props.niobe.context.replace(/<[^>]*>/g, '');
const introText = rawContext.length > 280
    ? rawContext.slice(0, 280) + '…'
    : rawContext.trim() || `Hi, I'm ${props.niobe.name}. Tell me what you'd like to order.`;
</script>

<template>
    <div class="relative min-h-screen overflow-hidden bg-[#0f0f12] px-6 py-10 text-[#e4e4e7]">
        <Head :title="`${niobe.name} – ${appName}`" />
        <div class="pointer-events-none absolute inset-0">
            <div class="absolute left-1/2 top-0 h-72 w-72 -translate-x-1/2 rounded-full bg-[#2563eb]/15 blur-3xl" />
        </div>

        <div class="relative mx-auto flex min-h-screen w-full max-w-3xl items-center justify-center">
            <div class="w-full rounded-2xl border border-white/10 bg-[#17171c] p-8 text-center shadow-2xl sm:p-10">
                <p class="text-xs uppercase tracking-[0.2em] text-[#71717a]">
                    AI waitress
                </p>
                <h1 class="mt-4 text-3xl font-semibold leading-tight text-white sm:text-4xl">
                    {{ niobe.name }}
                </h1>
                <p class="mx-auto mt-5 max-w-xl text-sm leading-7 text-[#a1a1aa] sm:text-base">
                    {{ introText }}
                </p>
                <p class="mt-6 text-sm text-[#71717a]">
                    Ready when you are.
                </p>

                <a
                    :href="talkHref"
                    class="mt-8 inline-flex items-center justify-center rounded-xl bg-[#3b82f6] px-6 py-3 text-sm font-medium text-white no-underline transition hover:bg-[#2563eb]"
                >
                    Talk to waitress
                </a>

                <p class="mt-8 text-xs uppercase tracking-[0.16em] text-[#52525b]">
                    {{ appName }}
                </p>
            </div>
        </div>
    </div>
</template>
