<script setup lang="ts">
import type { HTMLAttributes } from 'vue';

defineOptions({
    inheritAttrs: false,
});

type Props = {
    className?: HTMLAttributes['class'];
};

defineProps<Props>();
</script>

<template>
    <svg
        xmlns="http://www.w3.org/2000/svg"
        viewBox="0 0 24 24"
        fill="none"
        :class="className"
        v-bind="$attrs"
    >
        <path
            fill="#1a73e8"
            fill-rule="evenodd"
            clip-rule="evenodd"
            d="M5 3L7 3L17 21L19 21L19 3L17 3L7 21L5 21Z"
        />
    </svg>
</template>
