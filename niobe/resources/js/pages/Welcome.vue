<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import AppLogoIcon from '@/components/AppLogoIcon.vue';
import { dashboard, home, login, logout, register } from '@/routes';
import { index as waitressesIndex } from '@/routes/waitresses';

withDefaults(
    defineProps<{
        canRegister: boolean;
    }>(),
    { canRegister: true },
);
</script>

<template>
    <div class="flex min-h-screen flex-col bg-[#0f0f12] text-[#e4e4e7]">
        <Head title="Welcome" />
        <header class="border-b border-white/10 bg-[#18181b]">
            <div class="mx-auto flex h-14 max-w-4xl items-center justify-between gap-2 px-3 sm:gap-4 sm:px-6">
                <Link :href="home()" class="flex min-w-0 shrink items-center gap-2 text-[#e4e4e7] hover:opacity-80">
                    <AppLogoIcon class="h-7 w-7 shrink-0 fill-[#93c5fd] sm:h-8 sm:w-8" />
                    <span class="truncate text-base font-semibold sm:text-lg">{{ $page.props.name || 'Niobe' }}</span>
                </Link>
                <div class="flex shrink-0 flex-wrap items-center justify-end gap-2 sm:gap-4">
                <template v-if="$page.props.auth?.user">
                    <Link :href="dashboard()" class="inline-flex min-h-[44px] min-w-[44px] items-center text-sm font-medium text-[#a1a1aa] hover:text-[#e4e4e7] sm:min-h-0 sm:min-w-0 sm:py-2">Dashboard</Link>
                    <Link :href="waitressesIndex()" class="inline-flex min-h-[44px] min-w-[44px] items-center text-sm font-medium text-[#a1a1aa] hover:text-[#e4e4e7] sm:min-h-0 sm:min-w-0 sm:py-2">My Waitresses</Link>
                    <Link :href="logout.url()" method="post" as="button" class="inline-flex min-h-[44px] min-w-[44px] items-center text-sm font-medium text-[#a1a1aa] hover:text-[#e4e4e7] sm:min-h-0 sm:min-w-0 sm:py-2">Sign out</Link>
                </template>
                <template v-else>
                    <Link :href="login()" class="inline-flex min-h-[44px] min-w-[44px] items-center text-sm font-medium text-[#a1a1aa] hover:text-[#e4e4e7] sm:min-h-0 sm:min-w-0 sm:py-2">Sign in</Link>
                    <Link
                        v-if="canRegister"
                        :href="register()"
                        class="inline-flex min-h-[44px] items-center justify-center rounded-[4px] bg-[#3b82f6] px-4 py-2 text-sm font-semibold text-white hover:bg-[#2563eb]"
                    >
                        Create account
                    </Link>
                </template>
                </div>
            </div>
        </header>

        <main class="flex flex-1 items-center justify-center px-4 py-16 sm:py-24">
            <div class="w-full max-w-xl text-center">
                <p class="text-lg font-bold text-[#93c5fd] sm:text-xl">
                    Your AI Waitress
                </p>
                <p class="mb-8 text-base font-medium text-[#a1a1aa]">
                    Upload your menu. Share one link. Customers talk to your AI waitress by voice and place orders.
                </p>
                <template v-if="$page.props.auth?.user">
                    <Link
                        :href="waitressesIndex()"
                    class="inline-block rounded-[4px] bg-[#3b82f6] px-6 py-2.5 text-sm font-semibold text-white hover:bg-[#2563eb]"
                    >
                        Go to My Waitresses
                    </Link>
                </template>
                <Link
                    v-else
                    :href="register()"
                    class="inline-block rounded-[4px] bg-[#3b82f6] px-6 py-2.5 text-sm font-semibold text-white hover:bg-[#2563eb]"
                >
                    Get started
                </Link>
            </div>
        </main>

        <footer class="border-t border-white/10 bg-[#18181b] py-4">
            <p class="text-center text-xs font-medium text-[#71717a]">AI waitress for orders</p>
        </footer>
    </div>
</template>
