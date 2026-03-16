<script setup lang="ts">
import { Head, Link } from '@inertiajs/vue3';
import { Link2, Mic, Upload, UserPlus } from 'lucide-vue-next';
import AppLogoIcon from '@/components/AppLogoIcon.vue';
import { dashboard, home, login, logout, register } from '@/routes';
import { index as waitressesIndex } from '@/routes/waitresses';

withDefaults(
    defineProps<{
        canRegister: boolean;
    }>(),
    { canRegister: true },
);

const steps = [
    {
        icon: UserPlus,
        title: 'Create an account',
        text: 'Sign up in seconds. No credit card required.',
    },
    {
        icon: Upload,
        title: 'Upload your menu',
        text: 'Add photos of your menu. We extract dishes and prices automatically.',
    },
    {
        icon: Link2,
        title: 'Share one link',
        text: 'Send the link to customers—on a QR code, your site, or table tent.',
    },
    {
        icon: Mic,
        title: 'Customers talk & order',
        text: 'They speak to your AI waitress by voice and place orders. Orders go where you choose.',
    },
];
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

        <main class="relative flex-1 overflow-hidden">
            <!-- Soft ambient glow -->
            <div
                class="pointer-events-none absolute left-1/2 top-20 h-[420px] w-[420px] -translate-x-1/2 rounded-full bg-[#2563eb]/12 blur-[100px] transition-opacity duration-700"
                aria-hidden
            />
            <div
                class="pointer-events-none absolute bottom-1/4 right-0 h-64 w-64 rounded-full bg-[#93c5fd]/8 blur-[80px]"
                aria-hidden
            />

            <section class="relative px-4 pt-16 pb-12 sm:pt-20 sm:pb-16">
                <div class="mx-auto max-w-2xl text-center">
                    <p
                        class="animate-fade-in text-xs font-semibold uppercase tracking-[0.25em] text-[#93c5fd] sm:text-sm"
                        style="animation-delay: 0ms; animation-fill-mode: both;"
                    >
                        For restaurant owners
                    </p>
                    <h1 class="animate-fade-in mt-4 text-3xl font-bold leading-tight text-[#e4e4e7] sm:text-4xl md:text-5xl" style="animation-delay: 80ms; animation-fill-mode: both;">
                        One link. Customers talk to <span class="text-[#93c5fd]">your AI waitress</span> and order.
                    </h1>
                    <p class="animate-fade-in mx-auto mt-5 max-w-lg text-base leading-relaxed text-[#a1a1aa] sm:text-lg" style="animation-delay: 160ms; animation-fill-mode: both;">
                        Customers open the link, talk to your AI waitress by voice, and place orders. You choose where orders go—email, webhook, your Order Management system, etc
                    </p>
                    <div class="animate-fade-in mt-8 flex flex-wrap items-center justify-center gap-3" style="animation-delay: 240ms; animation-fill-mode: both;">
                        <template v-if="$page.props.auth?.user">
                            <Link
                                :href="waitressesIndex()"
                                class="inline-flex items-center justify-center rounded-[6px] bg-[#3b82f6] px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-[#3b82f6]/25 transition hover:bg-[#2563eb] hover:shadow-[#2563eb]/30"
                            >
                                Go to My Waitresses
                            </Link>
                        </template>
                        <Link
                            v-else
                            :href="register()"
                            class="inline-flex items-center justify-center rounded-[6px] bg-[#3b82f6] px-6 py-3 text-sm font-semibold text-white shadow-lg shadow-[#3b82f6]/25 transition hover:bg-[#2563eb] hover:shadow-[#2563eb]/30"
                        >
                            Get started free
                        </Link>
                    </div>
                </div>
            </section>

            <!-- How it works -->
            <section class="relative border-t border-white/10 bg-[#18181b]/60 px-4 py-14 sm:py-16">
                <div class="mx-auto max-w-4xl">
                    <h2 class="text-center text-sm font-semibold uppercase tracking-[0.2em] text-[#71717a]">
                        How it works
                    </h2>
                    <p class="mt-2 text-center text-xl font-semibold text-[#e4e4e7] sm:text-2xl">
                        Four steps to your AI waitress
                    </p>
                    <ul class="mt-10 grid gap-6 sm:grid-cols-2 lg:grid-cols-4">
                        <li
                            v-for="(step, i) in steps"
                            :key="step.title"
                            class="group relative rounded-xl border border-white/10 bg-[#0f0f12]/80 p-5 transition hover:border-[#3b82f6]/30 hover:bg-[#18181b]/80"
                            :style="{ animationDelay: `${320 + i * 60}ms` }"
                        >
                            <span
                                class="absolute -top-1.5 left-5 flex h-6 w-6 items-center justify-center rounded-full bg-[#3b82f6] text-xs font-bold text-white"
                                aria-hidden
                            >
                                {{ i + 1 }}
                            </span>
                            <component
                                :is="step.icon"
                                class="mt-2 h-8 w-8 text-[#93c5fd] transition group-hover:scale-105"
                                stroke-width="1.75"
                                aria-hidden
                            />
                            <h3 class="mt-4 text-base font-semibold text-[#e4e4e7]">
                                {{ step.title }}
                            </h3>
                            <p class="mt-1.5 text-sm leading-relaxed text-[#a1a1aa]">
                                {{ step.text }}
                            </p>
                        </li>
                    </ul>
                </div>
            </section>

            <!-- Bottom CTA -->
            <section class="relative border-t border-white/10 px-4 py-12 sm:py-16">
                <div class="mx-auto max-w-xl text-center">
                    <p class="text-lg font-semibold text-[#e4e4e7]">
                        Ready to let customers order by voice?
                    </p>
                    <p class="mt-2 text-sm text-[#a1a1aa]">
                        No setup fees. Create your first AI waitress in minutes.
                    </p>
                    <template v-if="!$page.props.auth?.user">
                        <Link
                            :href="register()"
                            class="mt-6 inline-flex items-center justify-center rounded-[6px] bg-[#3b82f6] px-6 py-3 text-sm font-semibold text-white transition hover:bg-[#2563eb]"
                        >
                            Create account
                        </Link>
                    </template>
                    <Link
                        v-else
                        :href="waitressesIndex()"
                        class="mt-6 inline-flex items-center justify-center rounded-[6px] bg-[#3b82f6] px-6 py-3 text-sm font-semibold text-white transition hover:bg-[#2563eb]"
                    >
                        Create a waitress
                    </Link>
                </div>
            </section>
        </main>

        <footer class="border-t border-white/10 bg-[#18181b] py-4">
            <p class="text-center text-xs font-medium text-[#71717a]">AI waitress for orders</p>
        </footer>
    </div>
</template>

<style scoped>
@keyframes fade-in {
    from {
        opacity: 0;
        transform: translateY(10px);
    }
    to {
        opacity: 1;
        transform: translateY(0);
    }
}
.animate-fade-in {
    animation: fade-in 0.6s ease-out forwards;
}
</style>
